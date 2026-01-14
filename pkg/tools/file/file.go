package file

import (
	"bufio"
	"fmt"
	"log"
	"novel-video-workflow/pkg/broadcast"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type FileManager struct {
	BroadcastService *broadcast.BroadcastService
}

func NewFileManager() *FileManager {
	return &FileManager{
		BroadcastService: broadcast.NewBroadcastService(),
	}
}

type ChapterStructure struct {
	ChapterDir  string
	TextFile    string
	AudioDir    string
	SubtitleDir string
	ImageDir    string
	SceneDir    string
}

// Comment 表示一个注释
type Comment struct {
	ID        string    `json:"id"`         // 注释唯一标识
	Content   string    `json:"content"`    // 注释内容
	Line      int       `json:"line"`       // 注释所在行号
	StartPos  int       `json:"start_pos"`  // 在行内的起始位置
	EndPos    int       `json:"end_pos"`    // 在行内的结束位置
	Type      string    `json:"type"`       // 注释类型 (info, warning, error, highlight等)
	CreatedAt time.Time `json:"created_at"` // 创建时间
	Author    string    `json:"author"`     // 注释作者
}

// CommentsCollection 存储文本的注释集合
type CommentsCollection struct {
	Filepath string    `json:"filepath"` // 关联的文件路径
	Comments []Comment `json:"comments"` // 注释列表
}

// ChapterContentMap 章节内容映射，键为章节号，值为章节内容
type ChapterContentMap map[int]string

var ChapterMap ChapterContentMap
var chapterMapMutex sync.Mutex // 保护ChapterMap的互斥锁

// 这里需要传递一个.txt的绝对路径
func (fm *FileManager) CreateInputChapterStructure(absDir string) (*ChapterStructure, error) {
	if c_map, err := fm.ExtractChapterTxt(absDir); err != nil {
		return nil, err
	} else {
		// 使用互斥锁保护ChapterMap的写入
		chapterMapMutex.Lock()
		ChapterMap = c_map
		chapterMapMutex.Unlock()

		// 循环c_map并创建文件夹，创建新的txt文本放到文件夹下
		for chapterNum, content := range c_map {
			fm.CreateChapterStructure(chapterNum, content, absDir)
		}
	}
	//构建input文件夹
	return nil, nil
}

// CreateChapterStructure 创建章节目录结构，格式为 chapter_XX/chapter_XX.txt
func (fm *FileManager) CreateChapterStructure(chapterNum int, content string, absDir string) error {
	// 获取基础目录路径
	basePath := filepath.Dir(absDir)

	// 格式化章节号，确保两位数格式（如 01, 02, ...）
	chapterFolderName := fmt.Sprintf("chapter_%02d", chapterNum)
	chapterFileName := fmt.Sprintf("chapter_%02d.txt", chapterNum)

	// 创建章节目录
	chapterDir := filepath.Join(basePath, chapterFolderName)
	err := os.MkdirAll(chapterDir, 0755)
	if err != nil {
		return fmt.Errorf("创建章节目录失败: %v", err)
	}

	// 创建章节文件
	filePath := filepath.Join(chapterDir, chapterFileName)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建章节文件失败: %v", err)
	}
	defer file.Close()

	// 写入内容
	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("写入章节内容失败: %v", err)
	}

	return nil
}

// ExtractChapterTxt 提取章节编号和对应的内容，返回章节编号到内容的映射
func (fm *FileManager) ExtractChapterTxt(fileDir string) (ChapterContentMap, error) {
	fileHandle, err := os.OpenFile(fileDir, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer fileHandle.Close()

	chapterMap := make(ChapterContentMap)
	var currentContent strings.Builder
	currentChapterFound := false
	var currentChapterNum int

	scanner := bufio.NewScanner(fileHandle)
	// 使用正则表达式匹配章节标记
	// 匹配以"第x章"、"第xx章"、"第xxx章"等开头的行
	re := regexp.MustCompile(`(?m)^\s*第[\p{N}\p{L}]+[章节][^\r\n]*$`)

	for scanner.Scan() {
		text := scanner.Text()

		// 检查当前行是否为章节标记
		if match := re.FindString(text); match != "" {
			// 如果已经找到了上一个章节的内容，保存它
			if currentChapterFound {
				chapterMap[currentChapterNum] = strings.TrimSpace(currentContent.String())
				currentContent.Reset()
			}

			// 提取章节数字
			numStr := strings.TrimPrefix(match, "第")
			numStr = strings.TrimSuffix(numStr, "章")
			numStr = strings.TrimSpace(numStr)

			// 转换为阿拉伯数字
			if atoi, err := strconv.Atoi(numStr); err != nil {
				currentChapterNum = fm.convertChineseNumberToArabic(numStr)
			} else {
				currentChapterNum = atoi
			}

			currentChapterFound = true

			// 将章节标题也加入内容中
			currentContent.WriteString(text)
			currentContent.WriteString("\n")
		} else {
			// 如果当前行不是章节标记，将其添加到当前内容中
			if currentChapterFound {
				currentContent.WriteString(text)
				currentContent.WriteString("\n")
			}
		}
	}

	// 处理最后一个章节的内容
	if currentChapterFound {
		chapterMap[currentChapterNum] = strings.TrimSpace(currentContent.String())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return chapterMap, nil
}

// ConvertChineseNumberToArabic 将中文数字转换为阿拉伯数字
func (fm *FileManager) convertChineseNumberToArabic(chineseNum string) int {
	chineseToArabic := map[string]int{
		// 基础数字
		"零": 0, "一": 1, "二": 2, "两": 2, "三": 3, "四": 4, "五": 5,
		"六": 6, "七": 7, "八": 8, "九": 9, "十": 10,
		// 十一到二十
		"十一": 11, "十二": 12, "十三": 13, "十四": 14, "十五": 15,
		"十六": 16, "十七": 17, "十八": 18, "十九": 19, "二十": 20,
		// 二十一到三十
		"二十一": 21, "二十二": 22, "二十三": 23, "二十四": 24, "二十五": 25,
		"二十六": 26, "二十七": 27, "二十八": 28, "二十九": 29, "三十": 30,
		// 三十一到四十
		"三十一": 31, "三十二": 32, "三十三": 33, "三十四": 34, "三十五": 35,
		"三十六": 36, "三十七": 37, "三十八": 38, "三十九": 39, "四十": 40,
		// 四十一到五十
		"四十一": 41, "四十二": 42, "四十三": 43, "四十四": 44, "四十五": 45,
		"四十六": 46, "四十七": 47, "四十八": 48, "四十九": 49, "五十": 50,
		// 五十一到六十
		"五十一": 51, "五十二": 52, "五十三": 53, "五十四": 54, "五十五": 55,
		"五十六": 56, "五十七": 57, "五十八": 58, "五十九": 59, "六十": 60,
		// 六十一到七十
		"六十一": 61, "六十二": 62, "六十三": 63, "六十四": 64, "六十五": 65,
		"六十六": 66, "六十七": 67, "六十八": 68, "六十九": 69, "七十": 70,
		// 七十一到八十
		"七十一": 71, "七十二": 72, "七十三": 73, "七十四": 74, "七十五": 75,
		"七十六": 76, "七十七": 77, "七十八": 78, "七十九": 79, "八十": 80,
		// 八十一到九十
		"八十一": 81, "八十二": 82, "八十三": 83, "八十四": 84, "八十五": 85,
		"八十六": 86, "八十七": 87, "八十八": 88, "八十九": 89, "九十": 90,
		// 九十一到九十九
		"九十一": 91, "九十二": 92, "九十三": 93, "九十四": 94, "九十五": 95,
		"九十六": 96, "九十七": 97, "九十八": 98, "九十九": 99,
	}
	if num, exists := chineseToArabic[chineseNum]; exists {
		return num
	}
	return 0
}

// output则参考input的结构生成目录结构，分出章节，每个章节内参考如下即可
/*
```
output/
└── 小说名称/
    └── chapter_01/
        ├── chapter_01.wav      # 音频文件
        ├── chapter_01.srt      # 字幕文件
        └── images/             # 图像目录
            ├── scene_01.png
            ├── scene_02.png
            └── ...
    └── chapter_02/
        ├── chapter_02.wav      # 音频文件
        ├── chapter_02.srt      # 字幕文件
        └── images/             # 图像目录
            ├── scene_01.png
            ├── scene_02.png
            └── ...
```
*/
func (fm *FileManager) CreateOutputChapterStructure(inpDir string) {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	//inpDir下的文件夹名字
	fold_name := ""
	items, err := os.ReadDir(inpDir)
	for _, item := range items {
		if item.IsDir() {
			fold_name = item.Name()
		}
	}
	if err != nil {
		log.Fatal(err)
	}
	// 1、创建这个文件夹
	os.Mkdir(filepath.Join(dir, "output", fold_name), os.ModePerm)

	// 创建子文件夹
	// 使用互斥锁保护ChapterMap的读取
	chapterMapMutex.Lock()
	defer chapterMapMutex.Unlock()

	for key, _ := range ChapterMap {
		f_name := fmt.Sprintf("chapter_%02d", key)
		//创建文件夹
		os.Mkdir(filepath.Join(dir, "output", fold_name, f_name), os.ModePerm)
	}
	fmt.Println(dir)
}
