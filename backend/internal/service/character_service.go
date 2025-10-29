package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"pixiv-tailor/backend/internal/models"
	"pixiv-tailor/backend/pkg/paths"
)

// CharacterService 角色特征服务
type CharacterService struct {
	characterDir string
}

// NewCharacterService 创建角色特征服务
func NewCharacterService() *CharacterService {
	pathManager := paths.GetPathManager()
	var charDir string
	if pathManager != nil {
		charDir = filepath.Join(pathManager.GetDataDir(), "characters")
	} else {
		charDir = "backend/data/characters"
	}

	// 确保目录存在
	os.MkdirAll(charDir, 0755)

	return &CharacterService{
		characterDir: charDir,
	}
}

// ExtractCharacterTags 从tag文件目录提取角色特征
func (s *CharacterService) ExtractCharacterTags(sourceDir string, request models.CharacterProfileRequest) (*models.ExtractCharacterTags, error) {
	// 读取源目录下的所有tag文件
	tagFiles, err := filepath.Glob(filepath.Join(sourceDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("读取tag文件失败: %v", err)
	}

	if len(tagFiles) == 0 {
		return nil, fmt.Errorf("目录中没有找到tag文件")
	}

	tagFrequency := make(map[string]int)
	tagWeights := make(map[string]float64)
	tagWeightsCount := make(map[string]int)

	// 统计每个tag文件
	for _, file := range tagFiles {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			continue
		}

		var tagData map[string]interface{}
		if err := json.Unmarshal(data, &tagData); err != nil {
			continue
		}

		// 提取tags_sorted
		if tagsSorted, ok := tagData["tags_sorted"].([]interface{}); ok {
			for _, tagObj := range tagsSorted {
				if tagMap, ok := tagObj.(map[string]interface{}); ok {
					name, _ := tagMap["Name"].(string)
					confidence, ok := tagMap["Confidence"].(float64)

					if !ok {
						continue
					}

					// 跳过要排除的标签
					shouldExclude := false
					for _, excludeTag := range request.ExcludeTags {
						if strings.Contains(strings.ToLower(name), strings.ToLower(excludeTag)) {
							shouldExclude = true
							break
						}
					}
					if shouldExclude {
						continue
					}

					// 统计频率和权重
					tagFrequency[name]++
					tagWeights[name] += confidence
					tagWeightsCount[name]++
				}
			}
		}
	}

	// 计算平均权重
	for tag := range tagWeights {
		if count, ok := tagWeightsCount[tag]; ok && count > 0 {
			tagWeights[tag] /= float64(count)
		}
	}

	// 应用过滤：移除 CategoryPerson 标签（如 1girl, solo）
	filteredFrequency := make(map[string]int)
	filteredWeights := make(map[string]float64)

	for tag, freq := range tagFrequency {
		// 跳过人物类别标签
		if ClassifyTag(tag) == CategoryPerson {
			continue
		}
		filteredFrequency[tag] = freq
		if weight, ok := tagWeights[tag]; ok {
			filteredWeights[tag] = weight
		}
	}

	return &models.ExtractCharacterTags{
		TagFrequency: filteredFrequency,
		TagWeights:   filteredWeights,
		TotalFiles:   len(tagFiles),
	}, nil
}

// CreateCharacterProfile 创建角色特征配置
func (s *CharacterService) CreateCharacterProfile(request models.CharacterProfileRequest) (*models.CharacterProfile, error) {
	// 提取特征
	tags, err := s.ExtractCharacterTags(request.SourceDir, request)
	if err != nil {
		return nil, err
	}

	// 根据频率和权重筛选关键标签
	finalTags := []string{}
	finalWeights := make(map[string]float64)

	// 首先添加核心关键词（如果有）
	for _, coreTag := range request.CoreTags {
		if weight, ok := tags.TagWeights[coreTag]; ok {
			finalTags = append(finalTags, coreTag)
			finalWeights[coreTag] = weight
		} else {
			// 即使tag文件中没有，也添加入核心关键词（使用默认权重）
			finalTags = append(finalTags, coreTag)
			finalWeights[coreTag] = 0.95 // 默认高权重
		}
	}

	// 然后添加强制包含的标签
	for _, includeTag := range request.IncludeTags {
		// 检查是否已经添加（避免重复）
		alreadyAdded := false
		for _, addedTag := range finalTags {
			if addedTag == includeTag {
				alreadyAdded = true
				break
			}
		}
		if alreadyAdded {
			continue
		}

		if weight, ok := tags.TagWeights[includeTag]; ok {
			finalTags = append(finalTags, includeTag)
			finalWeights[includeTag] = weight
		}
	}

	// 创建一个排序的tag列表
	type tagEntry struct {
		Name      string
		Category  TagCategory
		Frequency int
		Weight    float64
		Score     float64 // 综合得分
	}

	// 按类别分组标签
	characterTags := []tagEntry{}
	faceTags := []tagEntry{}
	hairTags := []tagEntry{}
	accessoryTags := []tagEntry{}
	personTags := []tagEntry{} // 新增：人物类别
	otherTags := []tagEntry{}

	for tag, freq := range tags.TagFrequency {
		// 跳过已经添加的标签（包括核心关键词和强制包含的标签）
		alreadyAdded := false

		// 检查是否在核心关键词中
		for _, coreTag := range request.CoreTags {
			if tag == coreTag {
				alreadyAdded = true
				break
			}
		}
		if alreadyAdded {
			continue
		}

		// 检查是否在强制包含的标签中
		for _, includeTag := range request.IncludeTags {
			if tag == includeTag {
				alreadyAdded = true
				break
			}
		}
		if alreadyAdded {
			continue
		}

		// 检查最小频率
		if freq < request.MinFrequency {
			continue
		}

		weight := tags.TagWeights[tag]
		// 检查最小权重
		if weight < request.MinWeight {
			continue
		}

		// 分类标签
		category := ClassifyTag(tag)

		// 计算综合得分 (频率权重40% + 置信度权重60%)
		score := float64(freq)/float64(tags.TotalFiles)*100*0.4 + weight*100*0.6

		entry := tagEntry{
			Name:      tag,
			Category:  category,
			Frequency: freq,
			Weight:    weight,
			Score:     score,
		}

		// 根据类别分类
		switch category {
		case CategoryCharacter:
			characterTags = append(characterTags, entry)
		case CategoryPerson:
			// 人物类别，不添加到最终结果
			personTags = append(personTags, entry)
			continue
		case CategoryFace:
			faceTags = append(faceTags, entry)
		case CategoryHair:
			hairTags = append(hairTags, entry)
		case CategoryAccessory:
			accessoryTags = append(accessoryTags, entry)
		case CategoryGeneral:
			// 跳过通用标签
			continue
		default:
			// 其他类别（body, clothing, action, background, quality, style 等）先存储起来
			otherTags = append(otherTags, entry)
		}
	}

	// 按得分排序各类别
	sort.Slice(characterTags, func(i, j int) bool {
		return characterTags[i].Score > characterTags[j].Score
	})
	sort.Slice(faceTags, func(i, j int) bool {
		return faceTags[i].Score > faceTags[j].Score
	})
	sort.Slice(hairTags, func(i, j int) bool {
		return hairTags[i].Score > hairTags[j].Score
	})
	sort.Slice(accessoryTags, func(i, j int) bool {
		return accessoryTags[i].Score > accessoryTags[j].Score
	})
	sort.Slice(otherTags, func(i, j int) bool {
		return otherTags[i].Score > otherTags[j].Score
	})

	// 优先添加角色核心标签（无限制）
	maxTags := request.MaxTags
	if maxTags <= 0 {
		maxTags = 30 // 默认30个标签
	}

	// 1. 首先添加所有角色核心标签（最重要）
	for _, entry := range characterTags {
		finalTags = append(finalTags, entry.Name)
		finalWeights[entry.Name] = entry.Weight
	}

	// 2. 然后添加面部特征标签（最多3个）
	faceCount := 0
	for _, entry := range faceTags {
		if faceCount >= 3 {
			break
		}
		finalTags = append(finalTags, entry.Name)
		finalWeights[entry.Name] = entry.Weight
		faceCount++
	}

	// 3. 添加头发标签（最多2个）
	hairCount := 0
	for _, entry := range hairTags {
		if hairCount >= 2 {
			break
		}
		finalTags = append(finalTags, entry.Name)
		finalWeights[entry.Name] = entry.Weight
		hairCount++
	}

	// 4. 添加配饰标签（最多2个）
	accessoryCount := 0
	for _, entry := range accessoryTags {
		if accessoryCount >= 2 {
			break
		}
		finalTags = append(finalTags, entry.Name)
		finalWeights[entry.Name] = entry.Weight
		accessoryCount++
	}

	// 5. 如果还有空间，添加其他高价值的标签（包括服装）
	for _, entry := range otherTags {
		if len(finalTags) >= maxTags {
			break
		}
		// 跳过无用类别（保留服装）
		if entry.Category == CategoryAction || entry.Category == CategoryBackground ||
			entry.Category == CategoryQuality || entry.Category == CategoryStyle {
			continue
		}
		finalTags = append(finalTags, entry.Name)
		finalWeights[entry.Name] = entry.Weight
	}

	// 创建角色配置
	profile := &models.CharacterProfile{
		ID:          time.Now().Unix(),
		Name:        request.Name,
		Description: request.Description,
		Tags:        finalTags,
		TagWeights:  finalWeights,
		SourceFiles: []string{}, // 不保存源文件列表
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存到文件
	if err := s.SaveCharacterProfile(profile); err != nil {
		return nil, err
	}

	return profile, nil
}

// SaveCharacterProfile 保存角色特征配置到文件
func (s *CharacterService) SaveCharacterProfile(profile *models.CharacterProfile) error {
	filename := filepath.Join(s.characterDir, fmt.Sprintf("%s.json", profile.Name))
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

// LoadCharacterProfile 加载角色特征配置
func (s *CharacterService) LoadCharacterProfile(name string) (*models.CharacterProfile, error) {
	filename := filepath.Join(s.characterDir, fmt.Sprintf("%s.json", name))
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var profile models.CharacterProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

// ListCharacterProfiles 列出所有角色配置文件
func (s *CharacterService) ListCharacterProfiles() ([]*models.CharacterProfile, error) {
	files, err := filepath.Glob(filepath.Join(s.characterDir, "*.json"))
	if err != nil {
		return nil, err
	}

	profiles := []*models.CharacterProfile{}
	for _, file := range files {
		_, filename := filepath.Split(file)
		name := strings.TrimSuffix(filename, ".json")
		profile, err := s.LoadCharacterProfile(name)
		if err != nil {
			continue
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// DeleteCharacterProfile 删除角色特征配置
func (s *CharacterService) DeleteCharacterProfile(name string) error {
	filename := filepath.Join(s.characterDir, fmt.Sprintf("%s.json", name))
	return os.Remove(filename)
}
