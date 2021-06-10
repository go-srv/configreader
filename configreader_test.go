package configreader

import (
	"bytes"
	"testing"
	"time"

	"github.com/spf13/afero"
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
	type MyStruct struct {
		Int int `default:"8"`
	}

	conf := MyStruct{}

	err := LoadDefault(&conf)
	assert.Nil(t, err)

	assert.Equal(t, int(8), conf.Int)
}

func TestRequired(t *testing.T) {
	defer testTearDown()

	type testConfig struct {
		Required    int `required:"true"`
		NotRequired int `required:"fasle"`
		NotSet      int
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
