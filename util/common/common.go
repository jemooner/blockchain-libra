package common

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"io"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

// GzipDecode gzip解压缩
func GzipDecode(byteBuffer *bytes.Buffer) ([]byte, error) {
	r, err := gzip.NewReader(byteBuffer)
	if err != nil {
		return nil, err
	}
	r.Close()

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// GzipEncode gzip压缩
func GzipEncode(byteData []byte) *bytes.Buffer {
	var b bytes.Buffer
	w, _ := gzip.NewWriterLevel(&b, 1)
	defer w.Close()
	w.Write(byteData)
	w.Flush()
	return &b
}

// IsEmpty 判读数据是否为空
func IsEmpty(a interface{}) bool {
	v := reflect.ValueOf(a)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.Interface() == reflect.Zero(v.Type()).Interface()
}

// GetRandString 随机生成N位字符串
func GetRandString(n int) string {
	mainBuff := make([]byte, n)
	_, err := io.ReadFull(rand.Reader, mainBuff)
	if err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(mainBuff)[:n]
}

// GetEntropyCSPRNG 生成随机指定位数byte
func GetEntropyCSPRNG(n int) []byte {
	mainBuff := make([]byte, n)
	_, err := io.ReadFull(rand.Reader, mainBuff)
	if err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
	return mainBuff
}

//生成32位md5字串
func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// PingHost 节点测试是否Ping通方法
func PingHost(host string) bool {
	d := net.Dialer{Timeout: time.Second * 10, LocalAddr: &net.TCPAddr{}}
	_, err := d.Dial("tcp", host)
	//defer conn.Close()
	if err != nil {
		return false
	}
	return true
}

// Regular 校验参数是否为正整数或浮点数
func Regular(data string) bool {
	pattern := `^\d+$ |^(\d+)(\.\d+)?$`
	reg := regexp.MustCompile(pattern)
	s := reg.FindAllStringSubmatch(data, -1)
	if len(s) != 0 {
		return true
	}
	return false
}

// Compress 压缩 使用gzip压缩成tar.gz
func Compress(files []*os.File, dest string) (err error) {

	fw, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer fw.Close()

	gw := gzip.NewWriter(fw)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, file := range files {
		name := file.Name()
		file.Close()
		fi, err := os.Stat(name)
		if err != nil {
			return err
		}

		// 信息头
		h := new(tar.Header)
		h.Name = fi.Name()
		h.Size = fi.Size()
		h.Mode = int64(fi.Mode())
		h.ModTime = fi.ModTime()

		// 写信息头
		err = tw.WriteHeader(h)
		if err != nil {
			return err
		}
		fs, err := os.Open(name)

		if err != nil {
			return err
		}

		if _, err = io.Copy(tw, fs); err != nil {
			return err
		}
		fs.Close()
	}

	return nil
}

// CreateFile 创建文件并写入指定内容
func CreateFile(fileName, data string) (*os.File, error) {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0766)
	if err != nil {
		return nil, err
		//DockerFile文件内容写入
	}
	file.WriteString(data)

	return file, nil
}

// CreateTarFile 根据传入的文件，创建指定的tar文件
func CreateTarFile(name string, files ...*os.File) (*os.File, error) {

	tarFile := make([]*os.File, len(files))
	for i, file := range files {
		tarFile[i] = file
	}

	err := Compress(tarFile, name)
	if err != nil {
		return nil, err
	}

	err = deleteFile(files...)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(name)

	if err != nil {
		return nil, err
	}
	return file, nil
}

// 删除文件，可传入多个
func deleteFile(files ...*os.File) error {
	for _, file := range files {
		file.Close()
		err := os.Remove(file.Name())
		if err != nil {
			return err
		}
	}
	return nil
}

// FileUpload 上传文件到指定URL
func FileUpload(URL string, file *os.File) (string, error) {
	resp, err := http.NewRequest("POST", URL, file)
	if err != nil {
		return "", err
	}

	resp.Header.Add("Content-Type", "application/tar")
	// 设置 TimeOut
	DefaultClient := http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
		},
	}

	res, err := DefaultClient.Do(resp)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	err = os.Remove(file.Name())
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

var Locker = make(map[string]*sync.RWMutex)

func Lock(index string) {
	for {
		_, ok := Locker[index]
		if !ok {
			Locker[index] = &sync.RWMutex{}
			break
		}
		//100ms轮训一次状态
		time.Sleep(100 * time.Millisecond)
	}

	Locker[index].Lock()
}

func Unlock(index string) {
	Locker[index].Unlock()
	//删除使用过的锁，避免map无限增加
	delete(Locker, index)
}

// 检查地址是否是有效的以太坊地址
func IsValidAddress(address interface{}) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	switch v := address.(type) {
	case string:
		return re.MatchString(v)
	case common.Address:
		return re.MatchString(v.Hex())
	default:
		return false
	}
}

// IsZeroAddress 验证它是否是一个0地址:0x0000000000000000000000000000000000000000
func IsZeroAddress(iaddress interface{}) bool {
	var address common.Address
	switch v := iaddress.(type) {
	case string:
		address = common.HexToAddress(v)
	case common.Address:
		address = v
	default:
		return false
	}

	zeroAddressBytes := common.FromHex("0x0000000000000000000000000000000000000000")
	addressBytes := address.Bytes()
	return reflect.DeepEqual(addressBytes, zeroAddressBytes)
}

// ToDecimal将wei（整数）转换为小数。 第二个参数是小数位数。
func ToDecimal(ivalue interface{}, decimals int) decimal.Decimal {
	value := new(big.Int)
	switch v := ivalue.(type) {
	case string:
		value.SetString(v, 10)
	case *big.Int:
		value = v
	}

	decimal.DivisionPrecision = 18
	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	num, _ := decimal.NewFromString(value.String())
	result := num.Div(mul)
	return result
}

//BigMulti 大数相乘
func BigMulti(a, b string) string {
	if a == "0" || b == "0" {
		return "0"
	}
	// string转换成[]byte，容易取得相应位上的具体值
	bsi := []byte(a)
	bsj := []byte(b)

	temp := make([]int, len(bsi)+len(bsj))
	//两数相乘，结果位数不会超过两乘数位数和，即temp的长度只可能为 len(num1)+len(num2) 或 len(num1)+len(num2)-1
	// 选最大的，免得位数不够
	for i := 0; i < len(bsi); i++ {
		for j := 0; j < len(bsj); j++ {
			// 对应每个位上的乘积，直接累加存入 temp 中相应的位置
			temp[i+j+1] += int(bsi[i]-'0') * int(bsj[j]-'0')
		}
	}

	//统一处理进位
	for i := len(temp) - 1; i > 0; i-- {
		temp[i-1] += temp[i] / 10 //对该结果进位（进到前一位）
		temp[i] = temp[i] % 10    //对个位数保留
	}

	// a 和 b 较小的时候，temp的首位为0
	// 为避免输出结果以0开头，需要去掉temp的0首位
	if temp[0] == 0 {
		temp = temp[1:]
	}
	//转换结果：将[]int类型的temp转成[]byte类型,
	//因为在未处理进位的情况下，temp每位的结果可能超过255(go中，byte类型实为uint8，最大为255),所以temp选用[]int类型
	//但在处理完进位后，不再会出现溢出
	res := make([]byte, len(temp)) //res 存放最终结果的ASCII码

	for i := 0; i < len(temp); i++ {
		res[i] = byte(temp[i] + '0')
	}

	return string(res)
}

// 根据燃气限价(单位)和燃气价格(单位)计算燃气成本
func CalcGasCost(gasLimit uint64, gasPrice *big.Int) *big.Int {
	gasLimitBig := big.NewInt(int64(gasLimit))
	return gasLimitBig.Mul(gasLimitBig, gasPrice)
}

// ToDecimal将wei（整数）转换为小数。 第二个参数是小数位数。
func GasDecimal(ivalue interface{}, decimals int) decimal.Decimal {
	value := new(big.Int)
	switch v := ivalue.(type) {
	case string:
		value.SetString(v, 10)
	case *big.Int:
		value = v
	}
	decimal.DivisionPrecision = 4
	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	num, _ := decimal.NewFromString(value.String())
	result := num.Div(mul)

	return result
}
