package configreader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/xerrors"
	"gopkg.in/yaml.v2"
)

// tags for parse the config struct tag annotations
const (
	tagKey        = "key"
	tagDefault    = "default"
	tagFlag       = "flag"
	tagEnv        = "env"
	tagRequired   = "required"
	tagValidation = "validation"
	skipKey       = "-"

	devEnv   = "dev"
	localEnv = "local"
)

var (
	// ErrNotStruct is returned when value passed to LoadConfig/ReadConfig is not a struct.
	ErrNotStruct = xerrors.New("value does not appear to be a struct")

	// ErrNotStructPointer is returned when value passed to LoadConfig/ReadConfig is not a pointer to a struct.
	ErrNotStructPointer = xerrors.New("value passed was not a struct pointer")
)

// ConfigReader wraps spf13/viper to read configs
type ConfigReader struct {
	viper   *viper.Viper
	flagset *pflag.FlagSet

	// config file name and paths to search for
	configName  string
	configPaths []string

	// The filesystem to read config from
	fs afero.Fs

	// EnvPrefix
	envPrefix string

	// Suffix to merge and override config files
	// override sequence: default <- env <- local
	fileEnvName string
	allowMerge  bool
}

var c *ConfigReader

func init() {
	c = New()
}

// New creates a ConfigReader instance
func New() *ConfigReader {
	c := new(ConfigReader)
	c.viper = viper.New()
	c.configName = "config"
	c.configPaths = []string{"."}
	c.fs = afero.NewOsFs()
	c.viper.SetFs(c.fs)

	c.envPrefix = "APP"

	c.allowMerge = true
	c.fileEnvName = "APP_ENV"

	c.viper.SetEnvPrefix(c.envPrefix)
	c.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	c.viper.AutomaticEnv()

	return c
}

// Reset the global ConfigReader instance, in the purpose for test
func Reset() {
	c = New()
}

// LoadConfig wraps the global ConfigReader instance
func LoadConfig(confPtr interface{}) error { return c.LoadConfig(confPtr) }

// LoadConfig loads configs from file and set values into the confPtr
func (c *ConfigReader) LoadConfig(confPtr interface{}) error {
	return c.loadConfig(confPtr)
}

// LoadFromFile loads configs by parsing the filepath and will load configs with several suffix
func LoadFromFile(filePath string, confPtr interface{}) error {
	return c.LoadFromFile(filePath, confPtr)
}

// LoadFromFile loads configs by parsing the filepath and will load configs with several suffix
func (c *ConfigReader) LoadFromFile(filePath string, confPtr interface{}) error {
	c.SetConfigFile(filePath)
	return c.loadConfig(confPtr)
}

// ReadConfig wraps the global ConfigReader instance
func ReadConfig(in io.Reader, confType string, confPtr interface{}) error {
	return c.ReadConfig(in, confType, confPtr)
}

// ReadConfig read configs from io.Reader and set values into confPtr
func (c *ConfigReader) ReadConfig(in io.Reader, confType string, confPtr interface{}) error {
	return c.readConfig(in, confType, confPtr)
}

// ReadFromFile wraps the global ConfigReader instance
func ReadFromFile(filePath string, confPtr interface{}) error {
	return c.ReadFromFile(filePath, confPtr)
}

// ReadFromFile read configs from the special file and set values into confPtr
func (c *ConfigReader) ReadFromFile(filename string, confPtr interface{}) error {
	file, err := afero.ReadFile(c.fs, filename)
	if err != nil {
		return err
	}

	ext := filepath.Ext(filename)
	if len(ext) <= 1 {
		return fmt.Errorf("filename: %s requires valid extension", filename)
	}
	configType := ext[1:]
	return c.ReadConfig(bytes.NewReader(file), configType, confPtr)
}

// SetConfigFile sets the config filename with path together
func SetConfigFile(filePath string) { c.SetConfigFile(filePath) }

// SetConfigFile sets the config filename with path together
func (c *ConfigReader) SetConfigFile(filePath string) {
	folder := filepath.Dir(filePath)
	filename := filepath.Base(filePath)
	basename := strings.TrimSuffix(filename, filepath.Ext(filename))

	c.AddConfigPath(folder)
	c.SetConfigName(basename)
}

// SetConfigName wraps the global ConfigReader instance
func SetConfigName(configName string) { c.SetConfigName(configName) }

// SetConfigName sets the config file name (exclude type suffix) to search for
func (c *ConfigReader) SetConfigName(configName string) {
	c.configName = configName
}

// SetConfigPaths wraps the global ConfigReader instance
func SetConfigPaths(paths []string) { c.SetConfigPaths(paths) }

// SetConfigPaths set a list of paths to search config files
func (c *ConfigReader) SetConfigPaths(paths []string) {
	c.configPaths = paths
}

// AddConfigPath wraps the global ConfigReader instance
func AddConfigPath(path string) { c.AddConfigPath(path) }

// AddConfigPath adds one path to search config files
func (c *ConfigReader) AddConfigPath(path string) {
	c.configPaths = append(c.configPaths, path)
}

// AllowMerge wraps the global ConfigReader instance
func AllowMerge(allow bool) { c.AllowMerge(allow) }

// AllowMerge allow ConfigReader to read env and local configs to override base configs
func (c *ConfigReader) AllowMerge(allow bool) {
	c.allowMerge = allow
}

// SetEnvName set the env name to be search when merge configs
func SetEnvName(env string) { c.SetEnvName(env) }

// SetEnvName set the env name to be search when merge configs
func (c *ConfigReader) SetEnvName(env string) {
	c.fileEnvName = strings.ToUpper(strings.ReplaceAll(env, ".", "_"))
}

// SetFs wraps the global ConfigReader instance
func SetFs(fs afero.Fs) { c.SetFs(fs) }

// SetFs set the filesystem to read config files from
func (c *ConfigReader) SetFs(fs afero.Fs) {
	if fs != nil {
		c.fs = fs
		c.viper.SetFs(fs)
	}
}

// SetEnvPrefix wraps the global ConfigReader instance
func SetEnvPrefix(in string) { c.SetEnvPrefix(in) }

// SetEnvPrefix sets the prefix of environment
// e.g. key is 'addr', prefix set to 'app', then env value is 'APP_ADDR'
func (c *ConfigReader) SetEnvPrefix(in string) {
	if in != "" {
		c.envPrefix = in
		c.viper.SetEnvPrefix(in)
	}
}

// SetFlagSet set the flagset to lookup
func SetFlagSet(flag *pflag.FlagSet) { c.SetFlagSet(flag) }

// SetFlagSet set the flagset to lookup
func (c *ConfigReader) SetFlagSet(flag *pflag.FlagSet) {
	c.flagset = flag
}

// Debug print viper values
func Debug() { c.Debug() }

// Debug print viper values
func (c *ConfigReader) Debug() {
	c.viper.Debug()
}

func PrintConfig(structPtr interface{}) {
	ref := reflect.ValueOf(structPtr).Elem()
	_ = walkThroughStruct("", ref, func(fieldKey string, structField reflect.StructField, structRef reflect.Value) error {
		fmt.Printf("%s: %v\n", fieldKey, structRef)
		return nil
	})
}

////////

// LoadDefault loads the default value if it have default annotation
// It's just a suger function that happens ConfigReader could load default
func LoadDefault(structPtr interface{}) error {
	err := checkStructPtr(structPtr)
	if err != nil {
		return err
	}

	ref := reflect.ValueOf(structPtr).Elem()

	return walkThroughStruct("", ref, loadDefault)
}

func loadDefault(fieldKey string, structField reflect.StructField, structRef reflect.Value) error {
	defaultValue := structField.Tag.Get(tagDefault)
	if defaultValue != "" {
		return populateStructField(structField, structRef, defaultValue)
	}

	return nil
}

////////

func (c *ConfigReader) loadConfig(confPtr interface{}) error {
	err := checkStructPtr(confPtr)
	if err != nil {
		return err
	}

	err = c.parseStructTags(confPtr)
	if err != nil {
		return err
	}

	err = c.loadConfigs()
	if err != nil {
		return err
	}

	err = c.checkValues(confPtr)
	if err != nil {
		return err
	}

	return c.populateStructValues(confPtr)
}

func (c *ConfigReader) loadConfigs() error {
	for _, configPath := range c.configPaths {
		c.viper.AddConfigPath(configPath)
	}

	c.viper.SetConfigName(c.configName)
	err := c.viper.ReadInConfig()
	if err != nil {
		return err
	}

	if c.allowMerge {
		var configFileNotFoundError viper.ConfigFileNotFoundError

		envSuffix := os.Getenv(c.fileEnvName)
		if len(envSuffix) <= 0 {
			envSuffix = devEnv
		}
		c.viper.SetConfigName(c.configName + "_" + envSuffix)
		err = c.viper.MergeInConfig()
		if err != nil && !xerrors.As(err, &configFileNotFoundError) {
			return err
		}

		c.viper.SetConfigName(c.configName + "_" + localEnv)
		err = c.viper.MergeInConfig()
		if err != nil && !xerrors.As(err, &configFileNotFoundError) {
			return err
		}
	}

	return nil
}

func (c *ConfigReader) readConfig(in io.Reader, configType string, confPtr interface{}) error {
	err := checkStructPtr(confPtr)
	if err != nil {
		return err
	}

	err = c.parseStructTags(confPtr)
	if err != nil {
		return err
	}

	c.viper.SetConfigType(configType)
	err = c.viper.ReadConfig(in)
	if err != nil {
		return err
	}

	err = c.checkValues(confPtr)
	if err != nil {
		return err
	}

	return c.populateStructValues(confPtr)
}

func checkStructPtr(confPtr interface{}) error {
	ptrRef := reflect.ValueOf(confPtr)
	if ptrRef.Kind() != reflect.Ptr {
		return ErrNotStructPointer
	}
	elemRef := ptrRef.Elem()
	if elemRef.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	return nil
}

////////

func (c *ConfigReader) parseStructTags(confPtr interface{}) error {
	ref := reflect.ValueOf(confPtr).Elem()

	return walkThroughStruct("", ref, c.parseStructTag)
}

func (c *ConfigReader) parseStructTag(fieldKey string, structField reflect.StructField, structRef reflect.Value) error {
	tag := structField.Tag
	c.bindDefaultValue(fieldKey, tag.Get(tagDefault))
	err := c.bindEnvValue(fieldKey, tag.Get(tagEnv))
	if err != nil {
		return err
	}
	return c.bindFlagValue(fieldKey, tag.Get(tagFlag), tag.Get(tagDefault))
}

func (c *ConfigReader) bindDefaultValue(fieldkey string, val string) {
	if val != "" {
		c.viper.SetDefault(fieldkey, val)
	}
}

func (c *ConfigReader) bindEnvValue(fieldkey string, envname string) error {
	if envname != "" {
		return c.viper.BindEnv(fieldkey, strings.ToUpper(envname))
	}
	return nil
}

func (c *ConfigReader) bindFlagValue(fieldkey string, flagname string, defval string) error {
	if flagname != "" {
		var flag *pflag.Flag

		// Not in the command, try search PFlags
		if c.flagset == nil {
			pflag.String(flagname, "", flagname)
			flag = pflag.Lookup(flagname)
		} else {
			flag = c.flagset.Lookup(flagname)
		}

		return c.viper.BindPFlag(fieldkey, flag)
	}
	return nil
}

////////// Check Values

func (c *ConfigReader) checkValues(confPtr interface{}) error {
	ref := reflect.ValueOf(confPtr).Elem()

	return walkThroughStruct("", ref, c.checkValueOfField)
}

func (c *ConfigReader) checkValueOfField(fieldKey string, structField reflect.StructField, structRef reflect.Value) error {
	checkFns := []FieldProcessor{
		c.checkRequiredValueOfField,
		c.validateValueOfField,
	}

	for _, check := range checkFns {
		if err := check(fieldKey, structField, structRef); err != nil {
			return err
		}
	}

	return nil
}

func (c *ConfigReader) checkRequiredValueOfField(fieldKey string, structField reflect.StructField, structRef reflect.Value) error {
	required := structField.Tag.Get(tagRequired) == "true"
	if required {
		if c.viper.Get(fieldKey) == nil {
			return fmt.Errorf("[%s] is required", fieldKey)
		}
	}

	return nil
}

////////// Value Validation

func (c *ConfigReader) validateValueOfField(fieldKey string, structField reflect.StructField, structRef reflect.Value) error {
	// If there is no value, ignore the validation check
	// If the value is required, it should use the require check
	val := c.viper.Get(fieldKey)
	if val == nil {
		return nil
	}

	validation := structField.Tag.Get(tagValidation)
	if len(validation) > 0 {
		splits := strings.Split(validation, ":")
		if len(splits) < 2 {
			return fmt.Errorf("invalid validation [%s] of key [%s]", validation, fieldKey)
		}

		action := splits[0]
		rules := splits[1]

		if action != "in" && action != "range" {
			return fmt.Errorf("unsupported action of validation [%s] of key [%s]", validation, fieldKey)
		}
		inAction := true
		if action == "range" {
			inAction = false
		}

		// Prepare for validation
		validationFailed := fmt.Errorf("[%s] did not pass validation. want [%s] real [%v]", fieldKey, validation, val)
		validationBadVal := fmt.Errorf("[%s] failed to parse validation values [%s]", fieldKey, validation)

		var ruleSplits []string
		var lAct, rAct, lValStr, rValStr string
		var err error

		if inAction {
			rules = strings.TrimSpace(rules)
			rules = strings.TrimLeft(rules, "[")
			rules = strings.TrimRight(rules, "]")
			ruleSplits = strings.Split(rules, ",")
		} else {
			lAct, rAct, lValStr, rValStr, err = parseRangeRule(rules)
			if err != nil {
				return err
			}
		}

		// let's start check by value types
		// Handle time.Duration as a special case
		typeName := structField.Type.String()
		switch typeName {
		case "time.Duration":
			val := c.viper.GetDuration(fieldKey).Nanoseconds()
			if inAction {
				ok := false
				for _, s := range ruleSplits {
					s = strings.TrimSpace(s)
					vd, err := time.ParseDuration(s)
					if err != nil {
						return validationBadVal
					}
					v := vd.Nanoseconds()

					if val == v {
						ok = true
						return nil
					}
				}
				if !ok {
					return validationFailed
				}
			} else {
				var lValDur, rValDur time.Duration
				var lVal, rVal int64
				if lAct != "inf" {
					lValDur, err = time.ParseDuration(lValStr)
					if err != nil {
						return validationBadVal
					}
					lVal = lValDur.Nanoseconds()
				}
				if rAct != "inf" {
					rValDur, err = time.ParseDuration(rValStr)
					if err != nil {
						return validationBadVal
					}
					rVal = rValDur.Nanoseconds()
				}

				if (lAct == ">=" && val < lVal) ||
					(lAct == ">" && val <= lVal) ||
					(rAct == "<=" && val > rVal) ||
					(rAct == "<" && val >= rVal) {
					return validationFailed
				}
				return nil
			}
		}

		// Handle base types
		switch structRef.Kind() {
		case reflect.String:
			if inAction {
				ok := false
				for _, s := range ruleSplits {
					s = strings.TrimSpace(s)
					if val == s {
						ok = true
						return nil
					}
				}
				if !ok {
					return validationFailed
				}
			} else {
				return fmt.Errorf("string does not support range validation [%s]", validation)
			}
		case reflect.Float32, reflect.Float64:
			val := c.viper.GetFloat64(fieldKey)
			if inAction {
				ok := false
				for _, s := range ruleSplits {
					s = strings.TrimSpace(s)
					v, err := strconv.ParseFloat(s, 64)
					if err != nil {
						return validationBadVal
					}

					if math.Abs(val-v) < 1e-3 {
						ok = true
						return nil
					}
				}
				if !ok {
					return validationFailed
				}
			} else {
				var lVal, rVal float64
				if lAct != "inf" {
					lVal, err = strconv.ParseFloat(lValStr, 64)
					if err != nil {
						return validationBadVal
					}
				}
				if rAct != "inf" {
					rVal, err = strconv.ParseFloat(rValStr, 64)
					if err != nil {
						return validationBadVal
					}
				}

				if (lAct == ">=" && val < lVal) ||
					(lAct == ">" && val <= lVal) ||
					(rAct == "<=" && val > rVal) ||
					(rAct == "<" && val >= rVal) {
					return validationFailed
				}
				return nil
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val := c.viper.GetInt64(fieldKey)
			if inAction {
				ok := false
				for _, s := range ruleSplits {
					s = strings.TrimSpace(s)
					v, err := strconv.ParseInt(s, 10, 64)
					if err != nil {
						return validationBadVal
					}

					if val == v {
						ok = true
						return nil
					}
				}
				if !ok {
					return validationFailed
				}
			} else {
				var lVal, rVal int64
				if lAct != "inf" {
					lVal, err = strconv.ParseInt(lValStr, 10, 64)
					if err != nil {
						return validationBadVal
					}
				}
				if rAct != "inf" {
					rVal, err = strconv.ParseInt(rValStr, 10, 64)
					if err != nil {
						return validationBadVal
					}
				}

				if (lAct == ">=" && val < lVal) ||
					(lAct == ">" && val <= lVal) ||
					(rAct == "<=" && val > rVal) ||
					(rAct == "<" && val >= rVal) {
					return validationFailed
				}
				return nil
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			val := c.viper.GetUint64(fieldKey)
			if inAction {
				ok := false
				for _, s := range ruleSplits {
					s = strings.TrimSpace(s)
					v, err := strconv.ParseUint(s, 10, 64)
					if err != nil {
						return validationBadVal
					}

					if val == v {
						ok = true
						return nil
					}
				}
				if !ok {
					return validationFailed
				}
			} else {
				var lVal, rVal uint64
				if lAct != "inf" {
					lVal, err = strconv.ParseUint(lValStr, 10, 64)
					if err != nil {
						return validationBadVal
					}
				}
				if rAct != "inf" {
					rVal, err = strconv.ParseUint(rValStr, 10, 64)
					if err != nil {
						return validationBadVal
					}
				}

				if (lAct == ">=" && val < lVal) ||
					(lAct == ">" && val <= lVal) ||
					(rAct == "<=" && val > rVal) ||
					(rAct == "<" && val >= rVal) {
					return validationFailed
				}
				return nil
			}
		default:
			return validationBadVal
		}
	}

	return nil
}

func parseRangeRule(rules string) (lAct, rAct, lVal, rVal string, err error) {
	invalidRule := fmt.Errorf("invalid range rule: [%s]", rules)

	rules = strings.TrimSpace(rules)

	switch rules[0] {
	case '[':
		lAct = ">="
	case '(':
		lAct = ">"
	default:
		err = invalidRule
		return
	}
	rules = rules[1:]

	last := len(rules) - 1
	switch rules[last] {
	case ']':
		rAct = "<="
	case ')':
		rAct = "<"
	default:
		err = invalidRule
		return
	}
	rules = rules[:last]

	splits := strings.Split(rules, ",")
	if len(splits) != 2 {
		err = invalidRule
		return
	}

	lVal = strings.TrimSpace(splits[0])
	rVal = strings.TrimSpace(splits[1])

	if len(lVal) == 0 {
		lAct = "inf"
	}
	if len(rVal) == 0 {
		rAct = "inf"
	}
	if lAct == "inf" && rAct == "inf" {
		err = invalidRule
		return
	}

	return
}

////////

func (c *ConfigReader) populateStructValues(confPtr interface{}) error {
	return c.viper.Unmarshal(confPtr, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = tagKey
	})
}

func populateStructField(field reflect.StructField, fieldValue reflect.Value, value string) error {
	typeName := field.Type.String()
	switch typeName {
	case "time.Duration":
		d, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("unable to parse duration (%s) to time.Duration for field: %s! Error: %s", value, field.Name, err.Error())
		}
		fieldValue.Set(reflect.ValueOf(d))
		return nil
	}

	switch fieldValue.Kind() {
	case reflect.String:
		if isZeroOfUnderlyingType(fieldValue.Interface()) {
			fieldValue.SetString(value)
		}

	case reflect.Bool:
		bvalue, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("unable to convert value (%s) to bool for field: %s! Error: %s", value, field.Name, err.Error())
		}

		if isZeroOfUnderlyingType(fieldValue.Interface()) {
			fieldValue.SetBool(bvalue)
		}

	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("unable to convert value (%s) to float for field: %s! Error: %s", value, field.Name, err.Error())
		}

		if isZeroOfUnderlyingType(fieldValue.Interface()) {
			fieldValue.SetFloat(floatValue)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("unable to convert value (%s) to int for field: %s! Error: %s", value, field.Name, err.Error())
		}

		if isZeroOfUnderlyingType(fieldValue.Interface()) {
			fieldValue.SetInt(intValue)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		intValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("unable to convert value (%s) to unsigned int for field: %s! Error: %s", value, field.Name, err.Error())
		}

		if isZeroOfUnderlyingType(fieldValue.Interface()) {
			fieldValue.SetUint(intValue)
		}
	case reflect.Slice, reflect.Array:
		ref := reflect.New(fieldValue.Type())
		ref.Elem().Set(reflect.MakeSlice(fieldValue.Type(), 0, 0))
		if value != "" && value != "[]" {
			if err := json.Unmarshal([]byte(value), ref.Interface()); err != nil {
				return err
			}
		}
		fieldValue.Set(ref.Elem().Convert(fieldValue.Type()))
	case reflect.Map:
		ref := reflect.New(fieldValue.Type())
		ref.Elem().Set(reflect.MakeMap(fieldValue.Type()))
		if value != "" && value != "{}" {
			if err := json.Unmarshal([]byte(value), ref.Interface()); err != nil {
				return err
			}
		}
		fieldValue.Set(ref.Elem().Convert(fieldValue.Type()))
	}
	return nil
}

////////

// FieldProcessor process one of the struct value, value means it is not a sub-struct
type FieldProcessor func(fullFieldKey string, structField reflect.StructField, structRef reflect.Value) error

func walkThroughStruct(rootKey string, structRef reflect.Value, processField FieldProcessor) error {
	structType := structRef.Type()
	for i := 0; i < structType.NumField(); i++ {
		currentField := structRef.Field(i)
		structField := structType.Field(i)
		tag := structField.Tag

		squash := structField.Type.Kind() == reflect.Struct && structField.Anonymous

		fieldKey := tag.Get(tagKey)
		// Deal with the "," in the key defines, not pass it during the walk through
		if strings.Contains(fieldKey, ",") {
			keys := strings.Split(fieldKey, ",")
			fieldKey = keys[0]

			if keys[1] == "squash" {
				squash = true
			}
		}
		if len(fieldKey) <= 0 && !squash {
			fieldKey = strings.ToLower(structField.Name)
		}

		if squash {
			fieldKey = ""
		}

		if fieldKey == skipKey {
			continue
		}

		fullFieldKey := ""
		if rootKey != "" {
			if fieldKey != "" {
				fullFieldKey = rootKey + "." + fieldKey
			} else {
				fullFieldKey = rootKey
			}
		} else {
			fullFieldKey = fieldKey
		}

		if ast.IsExported(structField.Name) {
			switch structField.Type.Kind() {
			case reflect.Struct:
				err := walkThroughStruct(fullFieldKey, currentField, processField)
				if err != nil {
					return err
				}
			default:
				err := processField(fullFieldKey, structField, currentField)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

////////

func isZeroOfUnderlyingType(x interface{}) bool {
	// Source: http://stackoverflow.com/questions/13901819/quick-way-to-detect-empty-values-via-reflection-in-go
	return x == reflect.Zero(reflect.TypeOf(x)).Interface()
}

////////
// For dump config to file

// DumpConfig wraps the global ConfigReader instance
func DumpConfig(filename string, confPtr interface{}) error {
	return c.DumpConfig(filename, confPtr)
}

// DumpConfig dumps the merged config to filepath
func (c *ConfigReader) DumpConfig(filename string, confPtr interface{}) error {
	var configType string

	ext := filepath.Ext(filename)
	if ext != "" {
		configType = ext[1:]
	} else {
		return fmt.Errorf("config type could not be determined for %s", filename)
	}

	if !stringInSlice(configType, SupportedExts) {
		return fmt.Errorf("config type [%s] is not supported", configType)
	}

	config := confPtr
	if config == nil {
		config = make(map[string]interface{})
	}

	force := false
	flags := os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	if !force {
		flags |= os.O_EXCL
	}
	filePermission := os.FileMode(0644)
	f, err := c.fs.OpenFile(filename, flags, filePermission)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := marshalWriter(f, configType, config); err != nil {
		return err
	}

	return f.Sync()
}

var SupportedExts = []string{"json", "yaml", "yml"}

func marshalWriter(f afero.File, configType string, config interface{}) error {
	switch configType {
	case "json":
		b, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return err
		}
		_, err = f.WriteString(string(b))
		if err != nil {
			return err
		}
	case "yaml", "yml":
		b, err := yaml.Marshal(config)
		if err != nil {
			return err
		}
		if _, err = f.WriteString(string(b)); err != nil {
			return err
		}
	}
	return nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if a == b {
			return true
		}
	}
	return false
}
