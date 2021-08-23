package configreader

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func testTearDown() {
	Reset()
}

func writeFile(fs afero.Fs, filepath string, data []byte) error {
	f, err := fs.Create(filepath)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return err
	}

	return f.Sync()
}

func TestDataTypes(t *testing.T) {
	defer testTearDown()

	type (
		MyInt    int
		MyInt8   int8
		MyInt16  int16
		MyInt32  int32
		MyInt64  int64
		MyUint   uint
		MyUint8  uint8
		MyUint16 uint16
		MyUint32 uint32
		MyUint64 uint64
		//MyUintptr uintptr
		MyFloat32 float32
		MyFloat64 float64
		MyBool    bool
		MyString  string
		MyMap     map[string]int
		MySlice   []int
	)

	type testConfig struct {
		Int    int
		Int8   int8
		Int16  int16
		Int32  int32
		Int64  int64
		Uint   uint
		Uint8  uint8
		Uint16 uint16
		Uint32 uint32
		Uint64 uint64
		//Uintptr   uintptr
		Float32   float32
		Float64   float64
		BoolTrue  bool
		BoolFalse bool
		String    string        `key:"str"`
		Duration  time.Duration `key:"dur"`

		Map   map[string]string
		Slice []string

		MyInt    MyInt
		MyInt8   MyInt8
		MyInt16  MyInt16
		MyInt32  MyInt32
		MyInt64  MyInt64
		MyUint   MyUint
		MyUint8  MyUint8
		MyUint16 MyUint16
		MyUint32 MyUint32
		MyUint64 MyUint64
		//MyUintptr   MyUintptr
		MyFloat32   MyFloat32
		MyFloat64   MyFloat64
		MyBoolTrue  MyBool
		MyBoolFalse MyBool
		MyString    MyString
		MyMap       MyMap
		MySlice     MySlice
	}

	configData := []byte(`{
		"int": "-333",
		"int8": "111",
		"int16": "16",
		"int32": "32",
		"int64": "64",
		"uint": "12",
		"uint8": "8",
		"uint16": "16",
		"uint32": "32",
		"uint64": "64",
		"uintptr": "1",
		"float32": "1.32",
		"float64": "1.64",
		"booltrue": "true",
		"boolfalse": "false",
		"str": "something",
		"myint": "23",
		"dur": "86400s",
        "map": {
            "k1": "v1",
            "k2": "v2"
        },
        "mymap": {
            "k1": "1",
            "k2": "2"
        },
        "slice": [
            "s1",
            "s2"
        ],
        "myslice": [
            "1",
            "2"
        ]
	}`)

	fs := afero.NewMemMapFs()
	err := writeFile(fs, "/tmp/config.json", configData)
	assert.Nil(t, err)

	SetFs(fs)
	SetConfigName("config")
	AddConfigPath("/tmp")

	conf := testConfig{}
	err = LoadConfig(&conf)
	assert.Nil(t, err)

	assert.Equal(t, int(-333), conf.Int)
	assert.Equal(t, int8(111), conf.Int8)
	assert.Equal(t, int16(16), conf.Int16)
	assert.Equal(t, int32(32), conf.Int32)
	assert.Equal(t, int64(64), conf.Int64)

	assert.Equal(t, uint(12), conf.Uint)
	assert.Equal(t, uint8(8), conf.Uint8)
	assert.Equal(t, uint16(16), conf.Uint16)
	assert.Equal(t, uint32(32), conf.Uint32)
	assert.Equal(t, uint64(64), conf.Uint64)
	//assert.Equal(t, uintptr(1), conf.Uintptr)

	assert.Equal(t, float32(1.32), conf.Float32)
	assert.Equal(t, float64(1.64), conf.Float64)

	assert.True(t, conf.BoolTrue)
	assert.False(t, conf.BoolFalse)

	assert.Equal(t, "something", conf.String)

	assert.Equal(t, MyInt(23), conf.MyInt)

	assert.Equal(t, time.Second*86400, conf.Duration)

	assert.Equal(t, 1, conf.MyMap["k1"])
	assert.Equal(t, 2, conf.MyMap["k2"])
	assert.Equal(t, "v1", conf.Map["k1"])
	assert.Equal(t, "v2", conf.Map["k2"])

	assert.Equal(t, 1, conf.MySlice[0])
	assert.Equal(t, 2, conf.MySlice[1])
	assert.Equal(t, "s1", conf.Slice[0])
	assert.Equal(t, "s2", conf.Slice[1])
}

func TestLoadDefault(t *testing.T) {
	type (
		MyInt     int
		MyInt8    int8
		MyInt16   int16
		MyInt32   int32
		MyInt64   int64
		MyUint    uint
		MyUint8   uint8
		MyUint16  uint16
		MyUint32  uint32
		MyUint64  uint64
		MyFloat32 float32
		MyFloat64 float64
		MyBool    bool
		MyString  string
	)

	type MyStruct struct {
		Int       int           `default:"1"`
		Int8      int8          `default:"8"`
		Int16     int16         `default:"16"`
		Int32     int32         `default:"32"`
		Int64     int64         `default:"64"`
		Uint      uint          `default:"1"`
		Uint8     uint8         `default:"8"`
		Uint16    uint16        `default:"16"`
		Uint32    uint32        `default:"32"`
		Uint64    uint64        `default:"64"`
		Float32   float32       `default:"32.32"`
		Float64   float64       `default:"64.64"`
		BoolTrue  bool          `default:"true"`
		BoolFalse bool          `default:"false"`
		String    string        `key:"str" default:"test"`
		Duration  time.Duration `key:"dur" default:"8s"`

		MyInt       MyInt     `default:"1"`
		MyInt8      MyInt8    `default:"8"`
		MyInt16     MyInt16   `default:"16"`
		MyInt32     MyInt32   `default:"32"`
		MyInt64     MyInt64   `default:"64"`
		MyUint      MyUint    `default:"1"`
		MyUint8     MyUint8   `default:"8"`
		MyUint16    MyUint16  `default:"16"`
		MyUint32    MyUint32  `default:"32"`
		MyUint64    MyUint64  `default:"64"`
		MyFloat32   MyFloat32 `default:"32.32"`
		MyFloat64   MyFloat64 `default:"64.64"`
		MyBoolTrue  MyBool    `default:"true"`
		MyBoolFalse MyBool    `default:"false"`
		MyString    MyString  `key:"str" default:"test"`
	}

	conf := MyStruct{}

	err := LoadDefault(&conf)
	assert.Nil(t, err)

	assert.Equal(t, int(1), conf.Int)
	assert.Equal(t, int8(8), conf.Int8)
	assert.Equal(t, int16(16), conf.Int16)
	assert.Equal(t, int32(32), conf.Int32)
	assert.Equal(t, int64(64), conf.Int64)

	assert.Equal(t, uint(1), conf.Uint)
	assert.Equal(t, uint8(8), conf.Uint8)
	assert.Equal(t, uint16(16), conf.Uint16)
	assert.Equal(t, uint32(32), conf.Uint32)
	assert.Equal(t, uint64(64), conf.Uint64)

	assert.Equal(t, float32(32.32), conf.Float32)
	assert.Equal(t, float64(64.64), conf.Float64)

	assert.True(t, conf.BoolTrue)
	assert.False(t, conf.BoolFalse)
	assert.Equal(t, "test", conf.String)
	assert.Equal(t, 8*time.Second, conf.Duration)

	assert.Equal(t, MyInt(1), conf.MyInt)
	assert.Equal(t, MyInt8(8), conf.MyInt8)
	assert.Equal(t, MyInt16(16), conf.MyInt16)
	assert.Equal(t, MyInt32(32), conf.MyInt32)
	assert.Equal(t, MyInt64(64), conf.MyInt64)

	assert.Equal(t, MyUint(1), conf.MyUint)
	assert.Equal(t, MyUint8(8), conf.MyUint8)
	assert.Equal(t, MyUint16(16), conf.MyUint16)
	assert.Equal(t, MyUint32(32), conf.MyUint32)
	assert.Equal(t, MyUint64(64), conf.MyUint64)

	assert.Equal(t, MyFloat32(32.32), conf.MyFloat32)
	assert.Equal(t, MyFloat64(64.64), conf.MyFloat64)

	assert.Equal(t, MyBool(true), conf.MyBoolTrue)
	assert.Equal(t, MyBool(false), conf.MyBoolFalse)
	assert.Equal(t, MyString("test"), conf.MyString)
}

func TestStructSliceDefault(t *testing.T) {
	// Skip this as currently seems viper only support string slice
	t.Skip()

	defer testTearDown()

	type TestConfig struct {
		K1 int `default:"1"`
		K2 int `default:"2"`
	}

	type TestBase struct {
		T []TestConfig
	}

	configData := []byte(`{
			"t": [
				{"k1": "111"}
			]
		}`)

	fs := afero.NewMemMapFs()
	err := writeFile(fs, "/tmp/config.json", configData)
	assert.Nil(t, err)

	SetFs(fs)
	SetConfigName("config")
	AddConfigPath("/tmp")

	conf := TestBase{}
	err = LoadConfig(&conf)
	assert.Nil(t, err)
	assert.Equal(t, 111, conf.T[0].K1)
	assert.Equal(t, 2, conf.T[0].K2)
}

func TestRequired(t *testing.T) {
	defer testTearDown()

	type testConfig struct {
		Required    int `required:"true"`
		NotRequired int `required:"fasle"`
		NotSet      int
		Defaults    int `default:"666" required:"true"`
	}

	configDataOK := []byte(`{
		"required": "-333",
		"notrequired": "111",
		"notset": "16"
	}`)

	configDataBad := []byte(`{
		"notrequired": "111",
		"notset": "16"
	}`)

	fs := afero.NewMemMapFs()
	err := writeFile(fs, "/tmp/configok.json", configDataOK)
	assert.Nil(t, err)

	SetFs(fs)
	SetConfigName("configok")
	AddConfigPath("/tmp")

	confok := testConfig{}
	err = LoadConfig(&confok)
	assert.Nil(t, err)

	Reset()
	fs = afero.NewMemMapFs()
	err = writeFile(fs, "/tmp/configbad.json", configDataBad)
	assert.Nil(t, err)

	SetFs(fs)
	SetConfigName("configbad")
	AddConfigPath("/tmp")

	confbad := testConfig{}
	err = LoadConfig(&confbad)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestSliceDefault(t *testing.T) {
	type MyStruct struct {
		IntSlice []int `default:"[8,10]"`
	}

	conf := MyStruct{}

	err := LoadDefault(&conf)
	assert.Nil(t, err)

	assert.Equal(t, int(8), conf.IntSlice[0])
	assert.Equal(t, int(10), conf.IntSlice[1])
}

func TestMapDefault(t *testing.T) {
	type MyStruct struct {
		MapS2I map[string]int    `default:"{\"k1\":1, \"k2\":2}"`
		MapS2S map[string]string `default:"{\"k1\":\"1\", \"k2\":\"2\"}"`
	}

	conf := MyStruct{}

	err := LoadDefault(&conf)
	assert.Nil(t, err)

	assert.Equal(t, int(1), conf.MapS2I["k1"])
	assert.Equal(t, int(2), conf.MapS2I["k2"])

	assert.Equal(t, "1", conf.MapS2S["k1"])
	assert.Equal(t, "2", conf.MapS2S["k2"])
}

func TestReadConfig(t *testing.T) {
	defer testTearDown()

	type testConfig struct {
		Int int
	}

	configData := []byte(`{
		"int": "-333"
	}`)

	conf := testConfig{}
	err := ReadConfig(bytes.NewBuffer(configData), "json", &conf)
	assert.Nil(t, err)

	assert.Equal(t, -333, conf.Int)
}

func TestDumpConfig(t *testing.T) {
	defer testTearDown()

	fs := afero.NewMemMapFs()

	SetFs(fs)

	type MyStruct struct {
		MapS2I map[string]int    `default:"{\"K1\":1, \"k2\":2}"`
		MapS2S map[string]string `default:"{\"K1\":\"V1\", \"k2\":\"2\"}"`
	}

	conf := MyStruct{}

	err := LoadDefault(&conf)
	assert.Nil(t, err)

	contentYaml := `maps2i:
  K1: 1
  k2: 2
maps2s:
  K1: V1
  k2: "2"
`
	yamlFile := "/tmp/test_config.yaml"
	err = DumpConfig(yamlFile, &conf)
	assert.Nil(t, err)

	bufYaml, err := afero.ReadFile(fs, yamlFile)
	assert.Nil(t, err)
	assert.Equal(t, contentYaml, string(bufYaml))

	ymlFile := "/tmp/test_config.yml"
	err = DumpConfig(ymlFile, &conf)
	assert.Nil(t, err)

	bufYml, err := afero.ReadFile(fs, ymlFile)
	assert.Nil(t, err)
	assert.Equal(t, contentYaml, string(bufYml))

	contentJson := `{
  "MapS2I": {
    "K1": 1,
    "k2": 2
  },
  "MapS2S": {
    "K1": "V1",
    "k2": "2"
  }
}`
	jsonFile := "/tmp/test_config.json"
	err = DumpConfig(jsonFile, &conf)
	assert.Nil(t, err)

	bufJson, err := afero.ReadFile(fs, jsonFile)
	assert.Nil(t, err)
	assert.Equal(t, contentJson, string(bufJson))

	tomlFile := "/tmp/test_config.toml"
	err = DumpConfig(tomlFile, &conf)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "[toml] is not supported")
}

func TestEnvValue(t *testing.T) {
	defer testTearDown()

	fs := afero.NewMemMapFs()

	SetFs(fs)

	configData := []byte(`{
		"k": "file_val"
	}`)
	err := writeFile(fs, "/tmp/config.json", configData)
	assert.Nil(t, err)

	type MyStruct struct {
		K string `env:"k" default:"default"`
	}

	os.Setenv("APP_K", "env_val")

	conf := MyStruct{}

	SetEnvPrefix("APP")
	AddConfigPath("/tmp")
	err = LoadConfig(&conf)
	assert.Nil(t, err)

	assert.Equal(t, "env_val", conf.K)
}

func TestFlagValue(t *testing.T) {
	defer testTearDown()

	fs := afero.NewMemMapFs()

	SetFs(fs)

	configData := []byte(`{
		"k": "file_val"
	}`)
	err := writeFile(fs, "/tmp/config.json", configData)
	assert.Nil(t, err)

	type MyStruct struct {
		K string `flag:"k" default:"default"`
	}

	conf := MyStruct{}

	flagSet := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flagSet.String("k", "flag_default_val", "")
	flagSet.Set("k", "flag_val")

	SetFlagSet(flagSet)
	AddConfigPath("/tmp")
	err = LoadConfig(&conf)
	assert.Nil(t, err)

	assert.Equal(t, "flag_val", conf.K)
}

func TestAnonymousNestedConfig(t *testing.T) {
	defer testTearDown()

	fs := afero.NewMemMapFs()

	SetFs(fs)

	configData := []byte(`{
		"base": "file_val"
	}`)
	err := writeFile(fs, "/tmp/config.json", configData)
	assert.Nil(t, err)

	type BaseStruct struct {
		Base string
	}

	type MyStruct struct {
		BaseStruct `key:",squash"`
	}

	conf := MyStruct{}

	AddConfigPath("/tmp")
	err = LoadConfig(&conf)
	assert.Nil(t, err)

	assert.Equal(t, "file_val", conf.Base)
}

func TestAnonymousNestedConfigRequired(t *testing.T) {
	defer testTearDown()

	fs := afero.NewMemMapFs()

	SetFs(fs)

	configData := []byte(`{
		"base": "file_val"
	}`)
	err := writeFile(fs, "/tmp/config.json", configData)
	assert.Nil(t, err)

	type BaseStruct struct {
		Base string `required:"true"`
	}

	type MyStruct struct {
		BaseStruct `key:",squash"`
	}

	conf := MyStruct{}

	AddConfigPath("/tmp")
	err = LoadConfig(&conf)
	assert.Nil(t, err)

	assert.Equal(t, "file_val", conf.Base)
}

func TestAnonymousNestedConfigRequiredNotSet(t *testing.T) {
	defer testTearDown()

	fs := afero.NewMemMapFs()

	SetFs(fs)

	configData := []byte(`{
		"base2": "file_val"
	}`)
	err := writeFile(fs, "/tmp/config.json", configData)
	assert.Nil(t, err)

	type BaseStruct struct {
		Base string `required:"true"`
	}

	type MyStruct struct {
		BaseStruct `key:",squash"`
	}

	conf := MyStruct{}

	AddConfigPath("/tmp")
	err = LoadConfig(&conf)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "required", "should throw error on required not set")

	assert.Equal(t, "", conf.Base)
}

func TestValidationInt(t *testing.T) {
	defer testTearDown()
	fs := afero.NewMemMapFs()

	SetFs(fs)

	configData := []byte(`{
		"ki": "2",
		"kr": "3"
	}`)
	filename := "/tmp/config.json"
	err := writeFile(fs, filename, configData)
	assert.Nil(t, err)

	type Conf struct {
		Ki    int `validation:"in:[2, 9]"`
		Kr    int `validation:"range:[2, 9]"`
		NoVal int `validation:"range:[2, 9]"`
	}

	conf := Conf{}

	SetConfigName("config")
	AddConfigPath("/tmp")

	err = LoadConfig(&conf)
	assert.Nil(t, err)
}

func TestValidationUint(t *testing.T) {
	defer testTearDown()
	fs := afero.NewMemMapFs()

	SetFs(fs)

	configData := []byte(`{
		"ki": "2",
		"kr": "3"
	}`)
	filename := "/tmp/config.json"
	err := writeFile(fs, filename, configData)
	assert.Nil(t, err)

	type Conf struct {
		Ki    uint `validation:"in:[2, 9]"`
		Kr    uint `validation:"range:[2, 9]"`
		NoVal uint `validation:"range:[2, 9]"`
	}

	conf := Conf{}

	SetConfigName("config")
	AddConfigPath("/tmp")

	err = LoadConfig(&conf)
	assert.Nil(t, err)
}

func TestValidationFloat(t *testing.T) {
	defer testTearDown()
	fs := afero.NewMemMapFs()

	SetFs(fs)

	configData := []byte(`{
		"ki": "2",
		"kr": "3"
	}`)
	filename := "/tmp/config.json"
	err := writeFile(fs, filename, configData)
	assert.Nil(t, err)

	type Conf struct {
		Ki    float32 `validation:"in:[2, 9]"`
		Kr    float32 `validation:"range:[2, 9]"`
		NoVal float32 `validation:"range:[2, 9]"`
	}

	conf := Conf{}

	SetConfigName("config")
	AddConfigPath("/tmp")

	err = LoadConfig(&conf)
	assert.Nil(t, err)
}

func TestValidationString(t *testing.T) {
	defer testTearDown()
	fs := afero.NewMemMapFs()

	SetFs(fs)

	configData := []byte(`{
		"ki": "2",
		"kr": "test"
	}`)
	filename := "/tmp/config.json"
	err := writeFile(fs, filename, configData)
	assert.Nil(t, err)

	type Conf struct {
		Ki    string `validation:"in:[2, 9]"`
		Kr    string `validation:"range:[2, 8]"`
		NoVal uint   `validation:"range:[2, 9]"`
	}

	conf := Conf{}

	SetConfigName("config")
	AddConfigPath("/tmp")

	err = LoadConfig(&conf)
	assert.Nil(t, err)
}

func TestValidationDuration(t *testing.T) {
	defer testTearDown()
	fs := afero.NewMemMapFs()

	SetFs(fs)

	configData := []byte(`{
		"ki": "2s",
		"kr": "3m"
	}`)
	filename := "/tmp/config.json"
	err := writeFile(fs, filename, configData)
	assert.Nil(t, err)

	type Conf struct {
		Ki    time.Duration `validation:"in:[2s, 9s]"`
		Kr    time.Duration `validation:"range:[2s, 9m]"`
		NoVal time.Duration `validation:"range:[2s, 9m]"`
	}

	conf := Conf{}

	SetConfigName("config")
	AddConfigPath("/tmp")

	err = LoadConfig(&conf)
	assert.Nil(t, err)
}
