package main

import (
	"os"

	"pixiv-tailor/backend/internal/commands"
	"pixiv-tailor/backend/internal/logger"
	"pixiv-tailor/backend/pkg/paths"

	"github.com/urfave/cli/v2"
)

const (
	version = "1.0.0"
)

func main() {
	// 初始化路径管理器
	if err := paths.InitPathManager(""); err != nil {
		logger.Errorf("初始化路径管理器失败: %v", err)
		os.Exit(1)
	}

	app := &cli.App{
		Name:    "pixiv-tailor",
		Usage:   "AI图像生成与训练工具",
		Version: version,
		Description: `PixivTailor 是一个基于Go语言开发的AI图像生成与训练工具集，
专门用于Pixiv插画的爬取、AI模型训练和图像生成。`,
		Authors: []*cli.Author{
			{
				Name:  "PixivTailor Team",
				Email: "team@pixivtailor.com",
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "启用详细输出",
				Action: func(ctx *cli.Context, v bool) error {
					logger.Init(v)
					if v {
						logger.SetLevel(logger.DebugLevel)
					} else {
						logger.SetLevel(logger.InfoLevel)
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "指定配置文件路径",
				Value:   "",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "server",
				Aliases: []string{"srv"},
				Usage:   "启动gRPC服务器",
				Description: `启动PixivTailor gRPC服务器，提供API服务。

示例:
  # 使用默认配置启动服务器
  pixiv-tailor server
  
  # 指定端口和配置
  pixiv-tailor server --port :8080 --config configs/custom.json
  
  # 指定数据库路径
  pixiv-tailor server --db data/custom.db`,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Usage:   "gRPC服务器端口",
						Value:   ":50051",
					},
					&cli.StringFlag{
						Name:    "db",
						Aliases: []string{"d"},
						Usage:   "数据库文件路径",
						Value:   "",
					},
				},
				Action: commands.ServerAction,
			},
			{
				Name:    "crawl",
				Aliases: []string{"c"},
				Usage:   "爬取Pixiv插画",
				Description: `从Pixiv爬取插画数据，支持按标签、用户或插画ID爬取。

示例:
  # 按标签爬取
  pixiv-tailor crawl --query "1girl" --limit 1000
  
  # 按用户爬取
  pixiv-tailor crawl --user-id 12345 --limit 50
  
  # 按插画ID爬取
  pixiv-tailor crawl --illust-id 12345678
  
  # 使用代理爬取
  pixiv-tailor crawl --query "1girl" --proxy --proxy-url "127.0.0.1:7890"
  
  # 使用自定义代理地址
  pixiv-tailor crawl --query "anime" --proxy --proxy-url "192.168.1.100:8080"`,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "query",
						Aliases:  []string{"q"},
						Usage:    "搜索查询",
						Required: false,
					},
					&cli.IntFlag{
						Name:    "user-id",
						Aliases: []string{"u"},
						Usage:   "用户ID",
					},
					&cli.IntFlag{
						Name:    "illust-id",
						Aliases: []string{"i"},
						Usage:   "插画ID",
					},
					&cli.StringFlag{
						Name:    "order",
						Aliases: []string{"o"},
						Usage:   "排序方式 (date_d/popular_d)",
						Value:   "date_d",
					},
					&cli.StringFlag{
						Name:    "mode",
						Aliases: []string{"m"},
						Usage:   "模式 (safe/r18/all)",
						Value:   "all",
					},
					&cli.IntFlag{
						Name:    "limit",
						Aliases: []string{"l"},
						Usage:   "最大下载数量",
						Value:   1000,
					},
					&cli.IntFlag{
						Name:    "delay",
						Aliases: []string{"d"},
						Usage:   "请求延迟(秒)",
						Value:   2,
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"O"},
						Usage:   "输出目录",
						Value:   "",
					},
					&cli.StringFlag{
						Name:    "save-name",
						Aliases: []string{"n"},
						Usage:   "保存名称",
					},
					&cli.BoolFlag{
						Name:    "proxy",
						Aliases: []string{"p"},
						Usage:   "启用代理",
						Value:   false,
					},
					&cli.StringFlag{
						Name:    "proxy-url",
						Aliases: []string{"proxy-addr"},
						Usage:   "代理地址 (格式: host:port)",
						Value:   "127.0.0.1:7890",
					},
				},
				Action: commands.CrawlAction,
			},
			{
				Name:    "generate",
				Aliases: []string{"g"},
				Usage:   "生成AI图像",
				Description: `使用Stable Diffusion生成AI图像，支持LoRA模型和姿态控制。

示例:
  # 基础生成
  pixiv-tailor generate --model "chosenMix_bakedVae.safetensors" --prompt "1girl, beautiful"
  
  # 使用LoRA模型
  pixiv-tailor generate --model "chosenMix_bakedVae.safetensors" --lora "spanking:0.8" --prompt "1girl"
  
  # 批量生成
  pixiv-tailor generate --model "chosenMix_bakedVae.safetensors" --batch-size 4 --prompt "1girl"`,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "model",
						Aliases:  []string{"m"},
						Usage:    "基础模型名称",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "prompt",
						Aliases:  []string{"p"},
						Usage:    "提示词",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "negative-prompt",
						Aliases: []string{"n"},
						Usage:   "负面提示词",
					},
					&cli.StringSliceFlag{
						Name:    "lora",
						Aliases: []string{"l"},
						Usage:   "LoRA模型 (格式: name:weight)",
					},
					&cli.Float64SliceFlag{
						Name:    "lora-weight",
						Aliases: []string{"w"},
						Usage:   "LoRA权重",
					},
					&cli.IntFlag{
						Name:    "batch-size",
						Aliases: []string{"b"},
						Usage:   "批次大小",
						Value:   1,
					},
					&cli.IntFlag{
						Name:    "steps",
						Aliases: []string{"s"},
						Usage:   "采样步数",
						Value:   20,
					},
					&cli.Float64Flag{
						Name:    "cfg-scale",
						Aliases: []string{"c"},
						Usage:   "CFG Scale",
						Value:   7.0,
					},
					&cli.IntFlag{
						Name:  "width",
						Usage: "图像宽度",
						Value: 512,
					},
					&cli.IntFlag{
						Name:  "height",
						Usage: "图像高度",
						Value: 512,
					},
					&cli.Int64Flag{
						Name:  "seed",
						Usage: "随机种子",
						Value: -1,
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "输出目录",
						Value:   "",
					},
				},
				Action: commands.GenerateAction,
			},
			{
				Name:    "train",
				Aliases: []string{"t"},
				Usage:   "训练LoRA模型",
				Description: `使用Kohya-ss训练LoRA模型。

示例:
  # 基础训练
  pixiv-tailor train --model-name "my_lora" --training-data "./data" --pretrained-model "chosenMix_bakedVae.safetensors"
  
  # 自定义参数训练
  pixiv-tailor train --model-name "my_lora" --training-data "./data" --epochs 10 --batch-size 2`,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "model-name",
						Aliases:  []string{"n"},
						Usage:    "模型名称",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "training-data",
						Aliases:  []string{"d"},
						Usage:    "训练数据目录",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "pretrained-model",
						Aliases: []string{"p"},
						Usage:   "预训练模型",
						Value:   "chosenMix_bakedVae.safetensors",
					},
					&cli.IntFlag{
						Name:    "epochs",
						Aliases: []string{"e"},
						Usage:   "训练轮数",
						Value:   5,
					},
					&cli.IntFlag{
						Name:    "batch-size",
						Aliases: []string{"b"},
						Usage:   "批次大小",
						Value:   1,
					},
					&cli.Float64Flag{
						Name:    "learning-rate",
						Aliases: []string{"r"},
						Usage:   "学习率",
						Value:   1e-4,
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "输出目录",
						Value:   "",
					},
				},
				Action: commands.TrainAction,
			},
			{
				Name:    "tag",
				Aliases: []string{"tag"},
				Usage:   "生成图像标签",
				Description: `使用WD14Tagger为图像生成标签。

示例:
  # 基础标签生成
  pixiv-tailor tag --input "./images" --output "./tagged"
  
  # 使用特定分析器
  pixiv-tailor tag --input "./images" --analyzer "wd14tagger" --threshold 0.5`,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "input",
						Aliases:  []string{"i"},
						Usage:    "输入目录",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "输出目录",
						Value:   "",
					},
					&cli.StringFlag{
						Name:    "analyzer",
						Aliases: []string{"a"},
						Usage:   "分析器 (wd14tagger/deepdanbooru)",
						Value:   "wd14tagger",
					},
					&cli.Float64Flag{
						Name:    "threshold",
						Aliases: []string{"t"},
						Usage:   "标签阈值",
						Value:   0.5,
					},
					&cli.StringFlag{
						Name:    "save-type",
						Aliases: []string{"s"},
						Usage:   "保存类型 (txt/json)",
						Value:   "txt",
					},
				},
				Action: commands.TagAction,
			},
			{
				Name:    "classify",
				Aliases: []string{"classify"},
				Usage:   "分类标签",
				Description: `使用OpenAI对标签进行分类。

示例:
  # 基础分类
  pixiv-tailor classify --input "tags.json" --output "classified.json"
  
  # 指定分类类别
  pixiv-tailor classify --input "tags.json" --categories "character,style,action"`,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "input",
						Aliases:  []string{"i"},
						Usage:    "输入文件",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "输出文件",
						Value:   "",
					},
					&cli.StringSliceFlag{
						Name:    "categories",
						Aliases: []string{"c"},
						Usage:   "分类类别",
						Value:   cli.NewStringSlice("character", "style", "action", "object"),
					},
				},
				Action: commands.ClassifyAction,
			},
		},
		Before: func(ctx *cli.Context) error {
			// 初始化日志系统
			verbose := ctx.Bool("verbose")
			logger.Init(verbose)
			if verbose {
				logger.SetLevel(logger.DebugLevel)
			} else {
				logger.SetLevel(logger.InfoLevel)
			}

			// 暂时跳过配置初始化
			// if ctx.Command.Name != "server" {
			// 	configPath := ctx.String("config")
			// 	if err := config.InitGlobalConfig(configPath); err != nil {
			// 		return fmt.Errorf("初始化配置失败: %v", err)
			// 	}
			// }

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Errorf("命令执行失败: %v", err)
		os.Exit(1)
	}
}
