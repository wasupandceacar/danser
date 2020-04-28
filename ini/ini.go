package ini

/**
 * 截入配置文件并读取其中内容
 * @Auther QiuXiangCheng
 * @Date   2018/05/08
 *
 * 与INI配置文件风格一样 根据顺序读取文件和每一行 如果在行首出现了;号，则认为是配置文件的注释
 * 当INI不规范时 如[mysql] 的注释被写为[mysql 则会返回错误
 * 本包的配置文件是严格区分大小写的 需要禁止区分大小写 将在后期加入或自行加入
 */

import (
	"bufio"
	"errors"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// 配置文件
type Config struct {
	conf map[string]url.Values
}

// 将指定的配置以字符串返回
func (c *Config) String(tag string) string {
	spl := strings.Split(tag, ".")
	key := strings.Join(spl[1:], "_")
	if len(spl) < 2 || spl[1] == "" {
		return ""
	}

	return c.conf[spl[0]].Get(key)
}

// 返回一个Int类型的配置值
func (c *Config) Int(tag string) (int, error) {
	return strconv.Atoi(c.String(tag))
}

// 返回一个int64配置值
func (c *Config) Int64(tag string) (int64, error) {
	return strconv.ParseInt(c.String(tag), 10, 64)
}

// 返回一个float64配置值
func (c *Config) Float64(tag string) (float64, error) {
	return strconv.ParseFloat(c.String(tag), 64)
}

// 初始化一个文件配置句柄
func NewFileConf(filePath string) (*Config, error) {

	cf := &Config{
		conf: make(map[string]url.Values, 10),
	}

	f, err := NewFileReader(filePath)
	if err != nil {
		return nil, errors.New("Error:can not read file \"" + filePath + "\"")
	}
	defer f.Close()

	tag := ""
	buf := bufio.NewReader(f)
	replacer := strings.NewReplacer(" ", "")

	for {
		lstr, err := buf.ReadString('\n')
		if err != nil && err != errors.New("EOF") {
			break
		}

		if lstr == "" {
			break
		}

		lstr = strings.TrimSpace(lstr)
		if lstr == "" {
			continue
		}

		if idx := strings.Index(lstr, "["); idx != -1 {
			if lstr[len(lstr)-1:] != "]" {
				return nil, errors.New("Error:field to parse this symbol style:\"" + lstr + "\"")
			}
			tag = lstr[1 : len(lstr)-1]
			cf.conf[tag] = url.Values{}
		} else {
			lstr = replacer.Replace(lstr)
			// 修改分隔符为 :
			spl := strings.Split(lstr, ":")

			// 增加注释符 //
			if lstr[0:1] == ";" || lstr[0:2] == "//" {
				continue
			}

			if len(spl) < 2 {
				return nil, errors.New("error:" + lstr)
			}
			cf.conf[tag].Set(strings.Replace(spl[0], ".", "_", -1), spl[1])
		}
	}

	return cf, nil
}

// 打开一个文件句柄
func NewFileReader(filePath string) (*os.File, error) {
	if !PathExists(filePath) {
		return nil, errors.New("Error:File not exists:" + filePath)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// 检查文件或文件夹是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	return false
}
