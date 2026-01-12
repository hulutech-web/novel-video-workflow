package drawthings

import (
	"os"
	"testing"

	"go.uber.org/zap"
)

func TestDrawThingsClient(t *testing.T) {
	// 创建logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// 创建客户端
	client := NewDrawThingsClient(logger, "http://localhost:7861")

	// 测试参数
	testPrompt := "A mysterious figure in a dark forest, suspenseful atmosphere"
	outputFile := "./test_output/test_image.png"

	// 确保输出目录存在
	os.MkdirAll("./test_output", 0755)

	// 测试文生图功能
	err = client.GenerateImageFromText(testPrompt, outputFile, 512, 512, true)
	if err != nil {
		t.Logf("Warning: Failed to generate image (expected if DrawThings API is not running): %v", err)
	} else {
		t.Logf("Successfully generated image: %s", outputFile)
		
		// 检查文件是否存在
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Errorf("Generated image file does not exist: %s", outputFile)
		} else {
			t.Logf("Generated image file exists: %s", outputFile)
		}
	}
}

func TestChapterImageGenerator(t *testing.T) {
	// 创建logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// 创建章节图像生成器
	generator := NewChapterImageGenerator(logger)

	// 测试章节文本
	chapterText := `深夜时分，古堡中传来阵阵奇怪的声音。走廊里回荡着脚步声，却不见一个人影。
	玛丽感到一阵寒意袭来，她紧紧抓住手中的手电筒。突然，墙上的油画似乎动了一下。`

	outputDir := "./test_output/chapter_images"
	
	// 确保输出目录存在
	os.MkdirAll(outputDir, 0755)

	// 测试章节图像生成
	results, err := generator.GenerateImagesFromChapter(chapterText, outputDir, 512, 512, true)
	if err != nil {
		t.Logf("Warning: Failed to generate chapter images (expected if DrawThings API is not running): %v", err)
	} else {
		t.Logf("Successfully generated %d images from chapter", len(results))
		for i, result := range results {
			t.Logf("Image %d: %s - %s", i+1, result.ImageFile, result.ParagraphText)
		}
	}
}