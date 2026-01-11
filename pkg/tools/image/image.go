/*图片生辰*/
package image

import (
	"fmt"
	"image/color"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/image/font"
)

// ImageGenerator 图片生成器
type ImageGenerator struct {
	logger *zap.Logger
}

// GeneratedImage 生成的图片信息
type GeneratedImage struct {
	ImageFile string `json:"image_file"`
	Prompt    string `json:"prompt"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}

// NewImageGenerator 创建图片生成器
func NewImageGenerator(logger *zap.Logger) *ImageGenerator {
	return &ImageGenerator{
		logger: logger,
	}
}

// GenerateSceneImages 根据文本生成场景图片
func (ig *ImageGenerator) GenerateSceneImages(text, outputDir string, maxImages int) ([]GeneratedImage, error) {
	// 提取场景描述词
	scenes := ig.extractSceneKeywords(text, maxImages)

	var results []GeneratedImage

	for i, scene := range scenes {
		if i >= maxImages {
			break
		}

		// 生成图片文件名
		imageFile := filepath.Join(outputDir, fmt.Sprintf("scene_%d.png", i+1))

		// 生成图片
		err := ig.generateImage(scene, imageFile)
		if err != nil {
			ig.logger.Warn("生成图片失败", zap.String("scene", scene), zap.Error(err))
			continue
		}

		results = append(results, GeneratedImage{
			ImageFile: imageFile,
			Prompt:    scene,
			Width:     1920,
			Height:    1080,
		})
	}

	return results, nil
}

// extractSceneKeywords 从文本中提取场景关键词
func (ig *ImageGenerator) extractSceneKeywords(text string, maxCount int) []string {
	// 简化的关键词提取 - 实际项目中可以使用更复杂的NLP技术
	keywords := []string{
		"自然风景", "城市景观", "室内场景", "森林", "海洋", "山脉", "夜景", "日出", "日落",
		"雨天", "雪景", "星空", "建筑", "人物", "动物", "花朵", "田野", "河流", "湖泊",
		"沙漠", "极光", "瀑布", "洞穴", "桥梁", "街道", "公园", "学校", "办公室", "家庭",
	}

	rand.Seed(time.Now().UnixNano())

	var selected []string
	used := make(map[string]bool)

	// 根据文本长度和复杂度生成相应数量的关键词
	textLen := len(text)
	sceneCount := min(maxCount, textLen/100+1)

	for len(selected) < sceneCount && len(selected) < len(keywords) {
		kw := keywords[rand.Intn(len(keywords))]
		if !used[kw] {
			selected = append(selected, kw)
			used[kw] = true
		}
	}

	if len(selected) == 0 {
		selected = []string{"自然风景", "室内场景"}
	}

	return selected
}

// generateImage 生成图片
func (ig *ImageGenerator) generateImage(prompt, outputFile string) error {
	width := 1920
	height := 1080

	dc := gg.NewContext(width, height)

	// 随机背景色
	bgColor := color.RGBA{
		R: uint8(rand.Intn(256)),
		G: uint8(rand.Intn(256)),
		B: uint8(rand.Intn(256)),
		A: 255,
	}
	dc.SetColor(bgColor)
	dc.Clear()

	// 随机绘制一些几何形状来模拟场景
	for i := 0; i < 20; i++ {
		x := rand.Float64() * float64(width)
		y := rand.Float64() * float64(height)
		radius := 20.0 + rand.Float64()*50.0

		fgColor := color.RGBA{
			R: uint8(rand.Intn(256)),
			G: uint8(rand.Intn(256)),
			B: uint8(rand.Intn(256)),
			A: 100,
		}
		dc.SetColor(fgColor)
		dc.DrawCircle(x, y, radius)
		dc.Fill()
	}

	// 绘制文字
	fontSize := 48.0
	dc.SetRGB(1, 1, 1) // 白色文字

	// 尝试加载字体
	var font font.Face
	fontPath := viper.GetString("image.font_path")
	if fontPath != "" {
		if fontBytes, err := os.ReadFile(fontPath); err == nil {
			if parsedFont, err := truetype.Parse(fontBytes); err == nil {
				font = truetype.NewFace(parsedFont, &truetype.Options{
					Size: fontSize,
				})
				dc.SetFontFace(font)
			}
		}
	}

	if font == nil {
		// 如果没有加载特定字体，使用默认字体
		dc.LoadFontFace("/System/Library/Fonts/PingFang.ttc", fontSize) // macOS
		if !strings.Contains(outputFile, "PingFang") {
			// 如果上面的字体加载失败，尝试其他常见字体路径
		}
	}

	// 计算文字尺寸并居中
	w, h := dc.MeasureString(prompt)
	x := (float64(width) - w) / 2
	y := (float64(height) - h) / 2

	// 添加文字阴影
	dc.SetRGB(0, 0, 0) // 黑色阴影
	dc.DrawString(prompt, x+2, y+2)
	dc.DrawString(prompt, x-2, y-2)
	dc.DrawString(prompt, x+2, y-2)
	dc.DrawString(prompt, x-2, y+2)

	// 添加主要文字
	dc.SetRGB(1, 1, 1) // 白色文字
	dc.DrawString(prompt, x, y)

	// 保存图片
	return dc.SavePNG(outputFile)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GenerateImageFromPrompt 根据提示生成图片
func (ig *ImageGenerator) GenerateImageFromPrompt(prompt, outputFile string) error {
	return ig.generateImage(prompt, outputFile)
}
