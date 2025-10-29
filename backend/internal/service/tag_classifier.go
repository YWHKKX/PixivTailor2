package service

import (
	"regexp"
	"strings"
)

// TagCategory 标签类别
type TagCategory string

const (
	CategoryCharacter  TagCategory = "character"  // 角色核心标签（最重要）
	CategoryPerson     TagCategory = "person"     // 人物描述（如 1girl, solo）
	CategoryFace       TagCategory = "face"       // 面部特征
	CategoryHair       TagCategory = "hair"       // 头发相关
	CategoryBody       TagCategory = "body"       // 身体特征
	CategoryClothing   TagCategory = "clothing"   // 服装
	CategoryAccessory  TagCategory = "accessory"  // 配饰
	CategoryAction     TagCategory = "action"     // 动作/姿势
	CategoryBackground TagCategory = "background" // 背景
	CategoryQuality    TagCategory = "quality"    // 质量标记
	CategoryStyle      TagCategory = "style"      // 风格
	CategoryGeneral    TagCategory = "general"    // 通用标签
)

// ClassifyTag 分类标签
func ClassifyTag(tag string) TagCategory {
	tagLower := strings.ToLower(tag)

	// 角色核心标签（包含 _(xxx), *_(series) 格式）
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+_\(.*\)$`, tag); matched {
		return CategoryCharacter
	}

	// 人物数量标签
	if tagLower == "1girl" || tagLower == "2girl" || tagLower == "3girl" ||
		tagLower == "1boy" || tagLower == "2boy" || tagLower == "3boy" ||
		tagLower == "solo" || tagLower == "multiple_girls" || tagLower == "multiple_boys" {
		return CategoryPerson
	}

	// 面部特征
	if strings.Contains(tagLower, "eye") || strings.Contains(tagLower, "look") ||
		strings.Contains(tagLower, "blush") || strings.Contains(tagLower, "smile") ||
		strings.Contains(tagLower, "mole") || strings.Contains(tagLower, "expression") {
		return CategoryFace
	}

	// 头发相关
	if strings.Contains(tagLower, "hair") || strings.Contains(tagLower, "bangs") ||
		strings.Contains(tagLower, "braid") || strings.Contains(tagLower, "ponytail") ||
		strings.Contains(tagLower, "twintails") || strings.Contains(tagLower, "ahoge") {
		return CategoryHair
	}

	// 身体特征
	if strings.Contains(tagLower, "breast") || strings.Contains(tagLower, "cleavage") ||
		strings.Contains(tagLower, "ass") || strings.Contains(tagLower, "thigh") ||
		strings.Contains(tagLower, "navel") || tagLower == "nude" || tagLower == "topless" {
		return CategoryBody
	}

	// 服装
	if strings.Contains(tagLower, "dress") || strings.Contains(tagLower, "skirt") ||
		strings.Contains(tagLower, "swimsuit") || strings.Contains(tagLower, "bikini") ||
		strings.Contains(tagLower, "shirt") || strings.Contains(tagLower, "jacket") ||
		strings.Contains(tagLower, "pants") || strings.Contains(tagLower, "shorts") ||
		strings.Contains(tagLower, "uniform") || strings.Contains(tagLower, "costume") ||
		tagLower == "pantyhose" || tagLower == "stockings" || tagLower == "socks" ||
		tagLower == "panties" || tagLower == "bra" || tagLower == "lingerie" {
		return CategoryClothing
	}

	// 配饰
	if strings.Contains(tagLower, "ornament") || strings.Contains(tagLower, "ribbon") ||
		strings.Contains(tagLower, "bow") || strings.Contains(tagLower, "jewelry") ||
		strings.Contains(tagLower, "necklace") || strings.Contains(tagLower, "earring") ||
		strings.Contains(tagLower, "glove") || strings.Contains(tagLower, "hat") ||
		strings.Contains(tagLower, "cape") || strings.Contains(tagLower, "cloak") ||
		strings.Contains(tagLower, "flower") {
		return CategoryAccessory
	}

	// 动作/姿势
	if strings.Contains(tagLower, "stand") || strings.Contains(tagLower, "sit") ||
		strings.Contains(tagLower, "lie") || strings.Contains(tagLower, "lying") ||
		strings.Contains(tagLower, "kneel") || strings.Contains(tagLower, "walk") ||
		strings.Contains(tagLower, "run") || strings.Contains(tagLower, "jump") ||
		tagLower == "squatting" || tagLower == "bent_over" || tagLower == "arms_up" {
		return CategoryAction
	}

	// 背景
	if strings.Contains(tagLower, "background") || strings.Contains(tagLower, "wallpaper") ||
		strings.Contains(tagLower, "indoors") || strings.Contains(tagLower, "outdoors") ||
		strings.Contains(tagLower, "sky") || strings.Contains(tagLower, "grass") ||
		strings.Contains(tagLower, "water") || strings.Contains(tagLower, "forest") ||
		strings.Contains(tagLower, "beach") || tagLower == "simple_background" {
		return CategoryBackground
	}

	// 质量标记
	if tagLower == "high_quality" || tagLower == "low_quality" ||
		tagLower == "masterpiece" || tagLower == "best_quality" ||
		strings.Contains(tagLower, "pixiv") || strings.Contains(tagLower, "rating") {
		return CategoryQuality
	}

	// 风格
	if strings.Contains(tagLower, "style") || strings.Contains(tagLower, "art") ||
		strings.Contains(tagLower, "anime") || strings.Contains(tagLower, "manga") ||
		tagLower == "realistic" || tagLower == "sketch" {
		return CategoryStyle
	}

	return CategoryGeneral
}
