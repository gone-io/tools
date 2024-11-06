package mock

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"errors"

	"github.com/spf13/cobra"
)

var scanDir string
var packageName string
var destinationDir string

var Command = &cobra.Command{
	Use:   "mock",
	Short: "generate mock goner code for interface",
	RunE: func(cmd *cobra.Command, args []string) error {
		if scanDir == "" {
			return errors.New("scanDir is required")
		}

		if packageName == "" {
			return errors.New("packageName is required")
		}

		if destinationDir == "" {
			destinationDir = path.Join(scanDir, packageName)
		}

		err := os.MkdirAll(destinationDir, os.ModePerm)
		if err != nil {
			return err
		}

		var mockedStructs []string

		// 使用 filepath.Walk 递归扫描目录
		err = filepath.Walk(scanDir, func(theFilepath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// 检查是否是文件（排除目录）
			if !info.IsDir() && !strings.HasSuffix(info.Name(), "gone.go") && strings.HasSuffix(info.Name(), ".go") {
				code, err := GenMockCode(theFilepath, packageName)
				if err != nil {
					return err
				}

				filename := filepath.Base(theFilepath)

				newFilepath := path.Join(destinationDir, fmt.Sprintf("%s.gone.go", strings.TrimSuffix(filename, ".go")))

				code = AddGoneCode(code)
				mocked := GetMockedInterface(code)
				mockedStructs = append(mockedStructs, mocked...)

				if len(mocked) > 0 {
					return os.WriteFile(newFilepath, []byte(code), 0644)
				}
			}
			return nil
		})
		if err != nil {
			return err
		}

		priestCode := GenMockPriestCode(mockedStructs, packageName)

		return os.WriteFile(path.Join(destinationDir, "priest.gone.go"), []byte(priestCode), 0644)
	},
}

func init() {
	Command.Flags().StringVarP(&scanDir, "scan-dir", "s", "", "scan dirs")
	Command.Flags().StringVarP(&packageName, "package", "p", "", "package name")
	Command.Flags().StringVarP(&destinationDir, "destination", "d", "", "destination dir")
}

func GenMockCode(source string, packageName string) (code string, err error) {
	cmd := exec.Command("mockgen",
		fmt.Sprintf("-source=%s", source),
		fmt.Sprintf("-package=%s", packageName),
		"-write_command_comment=false",
		"-write_package_comment=false",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

const mockGneTpl = "// Code generated by MockGen. DO NOT EDIT."
const GoneGenTpl = "// Code generated by Goner. DO NOT EDIT."

func AddGoneCode(code string) string {
	code = strings.Replace(code, mockGneTpl, GoneGenTpl, 1)

	code = strings.Replace(code, "import (", `import (
	"github.com/gone-io/gone"`, 1)
	return strings.Replace(code, "isgomock struct{}", "isgomock struct{}\n\tgone.Flag", -1)
}

var re = regexp.MustCompile(`(?s)type (\S*?) struct[^}]+?isgomock[^}]+?}`)

func GetMockedInterface(code string) (list []string) {
	//	使用正则表达式提取
	match := re.FindAllStringSubmatch(code, -1)

	for _, m := range match {
		list = append(list, m[1])
	}

	return list
}

// priest.gone.go
var priestTpl = `// Code generated by Goner. DO NOT EDIT.

package %s

import (
	"github.com/gone-io/gone"
	gomock "go.uber.org/mock/gomock"
)

func MockPriest(cemetery gone.Cemetery, ctrl *gomock.Controller) {
%s
}
`

func GenMockPriestCode(mockedStructs []string, packageName string) (code string) {
	var needBury []string
	for _, mocked := range mockedStructs {
		if strings.Contains(mocked, "[") {
			continue
		}
		needBury = append(needBury, fmt.Sprintf("\tcemetery.Bury(New%s(ctrl))", mocked))
	}
	return fmt.Sprintf(priestTpl, packageName, strings.Join(needBury, "\n"))
}
