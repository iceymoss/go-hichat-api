package sensitive

import (
	"fmt"

	"path/filepath"

	"github.com/importcjj/sensitive"
)

const SensitiveROOT = "resources/sensitive/"

type SensitiveType string

const (
	FILE_COVID_19        SensitiveType = "COVID-19词库.txt"
	FILE_GFW_SUPPLEMENT  SensitiveType = "GFW补充词库.txt"
	FILE_REACTIONARY     SensitiveType = "反动词库.txt"
	FILE_VIOLENCE_TERROR SensitiveType = "暴恐词库.txt"
	FILE_PORNOGRAPHY     SensitiveType = "色情词库.txt"
	FILE_SUPPLEMENT      SensitiveType = "补充词库.txt"
	FILE_TEMP_TENCENT    SensitiveType = "零时-Tencent.txt"
	FILE_OTHER           SensitiveType = "其他词库.txt"
	FILE_LIVELIHOOD      SensitiveType = "民生词库.txt"
	FILE_CORRUPTION      SensitiveType = "贪腐词库.txt"
	OTHER_FILE           SensitiveType = "other.txt"
	ALL_FILE             SensitiveType = "all_sensitive.txt"
)

type Word struct {
	Filter *sensitive.Filter
}

func NewWord(t SensitiveType) (*Word, error) {
	filter := sensitive.New()

	loadFile := string(SensitiveROOT + ALL_FILE)
	switch t {
	case FILE_COVID_19:
		loadFile = filepath.Join(SensitiveROOT, string(FILE_COVID_19))
	case FILE_GFW_SUPPLEMENT:
		loadFile = filepath.Join(SensitiveROOT, string(FILE_GFW_SUPPLEMENT))
	case FILE_REACTIONARY:
		loadFile = filepath.Join(SensitiveROOT, string(FILE_REACTIONARY))
	case FILE_VIOLENCE_TERROR:
		loadFile = filepath.Join(SensitiveROOT, string(FILE_VIOLENCE_TERROR))
	case FILE_PORNOGRAPHY:
		loadFile = filepath.Join(SensitiveROOT, string(FILE_PORNOGRAPHY))
	case FILE_SUPPLEMENT:
		loadFile = filepath.Join(SensitiveROOT, string(FILE_SUPPLEMENT))
	case FILE_TEMP_TENCENT:
		loadFile = filepath.Join(SensitiveROOT, string(FILE_TEMP_TENCENT))
	case FILE_OTHER:
		loadFile = filepath.Join(SensitiveROOT, string(FILE_OTHER))
	case FILE_LIVELIHOOD:
		loadFile = filepath.Join(SensitiveROOT, string(FILE_LIVELIHOOD))
	case FILE_CORRUPTION:
		loadFile = filepath.Join(SensitiveROOT, string(FILE_CORRUPTION))
	case ALL_FILE:
		loadFile = filepath.Join(SensitiveROOT, string(ALL_FILE))
	case OTHER_FILE:
		loadFile = filepath.Join(SensitiveROOT, string(OTHER_FILE))
	default:
		return nil, fmt.Errorf("未知的敏感词类型: %s", t)
	}

	err := filter.LoadWordDict(loadFile)
	if err != nil {
		return nil, err
	}

	return &Word{
		Filter: filter,
	}, nil
}

func (w *Word) Validate(content string) (bool, string) {
	return w.Filter.Validate(content)
}

func (w *Word) Replace(content string, replChar rune) string {
	return w.Filter.Replace(content, replChar)
}
