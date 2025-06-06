package env

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"denv/internal/config"
	"denv/internal/filehandler"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"gopkg.in/yaml.v3"
)

/*
 * Protocol:
 *
 * ```
 * id: my-key
 * ---
 * key1: value1
 * key2: value2
 * ---
 * Any other payload
 * ```
 */

type DynamicEnvMetadata struct {
	ID string `yaml:"id"`
}
type DynamicEnvValue struct {
	Metadata DynamicEnvMetadata
	Raw      string
	Data     map[string]any
	Payload  string
}

type DynamicEnv struct {
	Config      *config.ConfigType
	UserConfig  *config.UserConfigType
	Filehandler *filehandler.FileHandler
	index       *map[string]string
}

type DynamicEnvParsed struct {
	Local map[string]string
	Env   map[string]string
}

func NewDynamicEnv(config *config.ConfigType, userConfig *config.UserConfigType, filehandler *filehandler.FileHandler) *DynamicEnv {
	return &DynamicEnv{Config: config, UserConfig: userConfig, Filehandler: filehandler}
}

func (d *DynamicEnv) GetFilePath(path string) string {
	if strings.HasPrefix(path, d.Config.DataDir+"/") {
		return path + d.Config.EnvSuffix
	}
	return path
}

func (d *DynamicEnv) ParseRawValue(value string, includeMetadata bool) (*DynamicEnvValue, error) {
	lines := strings.Split(value, "\n")
	var metadata DynamicEnvMetadata

	i := -1
	if includeMetadata {
		i = indexOf(lines, "---", 0)
		if i < 0 {
			return nil, errors.New("invalid YAML content: missing metadata separator")
		}

		rawMetadata := make(map[string]any)
		if err := yaml.Unmarshal([]byte(strings.Join(lines[:i], "\n")), &rawMetadata); err != nil {
			return nil, errors.New("invalid metadata: " + err.Error())
		}

		id, ok := rawMetadata["id"].(string)
		if !ok || id == "" {
			return nil, errors.New("invalid metadata: missing or invalid 'id'")
		}

		metadata.ID = id
	}

	j := indexOf(lines, "---", i+1)
	if j < 0 {
		j = len(lines)
	}

	raw := strings.Join(lines[i+1:j], "\n")
	data := make(map[string]any)
	if err := yaml.Unmarshal([]byte(raw), &data); err != nil {
		return nil, errors.New("invalid data: " + err.Error())
	}

	payload := ""
	if j < len(lines) {
		payload = strings.Join(lines[j+1:], "\n")
	}

	return &DynamicEnvValue{
		Metadata: metadata,
		Raw:      raw,
		Data:     data,
		Payload:  payload,
	}, nil
}

func indexOf(lines []string, target string, offset int) int {
	for i := offset; i < len(lines); i++ {
		if lines[i] == target {
			return i
		}
	}
	return -1
}

func (d *DynamicEnv) FormatValue(value *DynamicEnvValue, includeMetadata bool) (string, error) {
	if value == nil {
		return "", errors.New("value is nil")
	}

	output := ""
	if includeMetadata {
		data, err := yaml.Marshal(value.Metadata)
		if err != nil {
			return "", errors.New("failed to marshal metadata: " + err.Error())
		}
		output += string(data) + "\n---\n"
	}

	output += value.Raw
	if value.Payload != "" {
		output += "\n---\n" + value.Payload
	}

	return output, nil
}

func (d *DynamicEnv) EncryptData(data string) (string, error) {
	if len(d.UserConfig.Data.Recipients) == 0 {
		return "", errors.New("no recipient is added")
	}

	args := []string{"-a"}
	for _, recipient := range d.UserConfig.Data.Recipients {
		args = append(args, "-r", recipient)
	}

	cmd := exec.Command("age", args...)
	cmd.Stdin = strings.NewReader(data)

	output, err := cmd.Output()
	if err != nil {
		return "", errors.New("failed to encrypt data: " + err.Error())
	}

	return string(output), nil
}

func (d *DynamicEnv) DecryptData(data string) (string, error) {
	if d.Config.Identities == "" {
		return "", errors.New("no identities file provided")
	}

	args := []string{"--decrypt", "-i", d.Config.Identities}

	cmd := exec.Command("age", args...)
	cmd.Stdin = strings.NewReader(data)

	output, err := cmd.Output()
	if err != nil {
		return "", errors.New("failed to decrypt data: " + err.Error())
	}

	return string(output), nil
}

func (d *DynamicEnv) LoadValue(encrypted string) (*DynamicEnvValue, error) {
	value, err := d.DecryptData(encrypted)
	if err != nil {
		return nil, errors.New("failed to decrypt data: " + err.Error())
	}
	dynamicEnvValue, err := d.ParseRawValue(value, true)
	return dynamicEnvValue, err
}

func (d *DynamicEnv) ListEnvFiles(prefix string) ([]string, error) {
	files, err := d.Filehandler.ListFiles(filepath.Join(d.Config.DataDir, prefix), d.Config.DataDir)
	if err != nil {
		return nil, err
	}

	result := make([]string, len(files))
	for _, file := range files {
		file = strings.ReplaceAll(file, "\\", "/")
		result = append(result, file)
	}
	return result, nil
}

func (d *DynamicEnv) ListItems(prefix string) map[string]*DynamicEnvValue {
	envs := make(map[string]*DynamicEnvValue)
	files, err := d.ListEnvFiles(prefix)
	if err != nil {
		if d.Config.Debug {
			log.Printf("Error listing files: %v\n", err)
		}
		return envs
	}

	for _, file := range files {
		if !strings.HasSuffix(file, d.Config.EnvSuffix) {
			continue
		}
		value, err := d.Filehandler.ReadFile(filepath.Join(d.Config.DataDir, file))
		if err != nil {
			if d.Config.Debug {
				log.Printf("Error reading file %s: %v\n", file, err)
			}
			continue
		}
		dynamicEnvValue, err := d.LoadValue(value)
		if err != nil {
			if d.Config.Debug {
				log.Printf("Error parsing file %s: %v\n", file, err)
			}
			continue
		}
		uid := strings.TrimSuffix(file, d.Config.EnvSuffix)
		envs[uid] = dynamicEnvValue
	}
	return envs
}

func (d *DynamicEnv) LoadIndex() *map[string]string {
	if d.index != nil {
		return d.index
	}
	index := make(map[string]string)
	d.index = &index
	data, err := d.Filehandler.ReadFile(d.Config.IndexFile)
	if err != nil {
		return d.index
	}
	if err := yaml.Unmarshal([]byte(data), &index); err != nil {
		return d.index
	}
	return d.index
}

func (d *DynamicEnv) SaveIndex(index *map[string]string) error {
	indexContent, err := yaml.Marshal(index)
	if err != nil {
		return err
	}
	d.index = index
	return d.Filehandler.WriteFile(d.Config.IndexFile, string(indexContent))
}

func (d *DynamicEnv) BuildIndex() error {
	envs := d.ListItems("")
	index := make(map[string]string)
	for uid, value := range envs {
		index[uid] = value.Metadata.ID
	}
	return d.SaveIndex(&index)
}

func (d *DynamicEnv) UpdateIndex(uid string, id string, idFrom string) error {
	index := d.LoadIndex()
	if id == "" {
		delete(*index, id)
	} else {
		(*index)[uid] = id
	}
	return d.SaveIndex(d.index)
}

func (d *DynamicEnv) GetEnvUID(key string) (string, error) {
	index := d.LoadIndex()
	uid := ""
	for iUid, iId := range *index {
		if iId == key {
			uid = iUid
			break
		}
	}
	if uid == "" {
		nid, err := gonanoid.New()
		if err != nil {
			return "", errors.New("failed to generate ID: " + err.Error())
		}
		uid = nid
	}
	return uid, nil
}

func (d *DynamicEnv) GetEnvPath(uid string) string {
	path := filepath.Join(d.Config.DataDir, uid+d.Config.EnvSuffix)
	return path
}

func (d *DynamicEnv) GetEnv(key string) (*DynamicEnvValue, error) {
	uid, err := d.GetEnvUID(key)
	if err != nil {
		return nil, err
	}
	path := d.GetEnvPath(uid)
	data, err := d.Filehandler.ReadFile(path)
	if err != nil {
		return nil, err
	}
	dynamicEnvValue, err := d.LoadValue(data)
	return dynamicEnvValue, err
}

func (d *DynamicEnv) SetEnv(key string, value *DynamicEnvValue) error {
	if value == nil {
		return errors.New("value is nil")
	}

	keyFrom := value.Metadata.ID
	value.Metadata.ID = key

	uid, err := d.GetEnvUID(keyFrom)
	if err != nil {
		return err
	}

	data, err := d.FormatValue(value, true)
	if err != nil {
		return err
	}

	encrypted, err := d.EncryptData(data)
	if err != nil {
		return err
	}

	path := d.GetEnvPath(uid)
	if err := d.Filehandler.WriteFile(path, encrypted); err != nil {
		return err
	}

	return d.UpdateIndex(uid, key, keyFrom)
}

func (d *DynamicEnv) DeleteEnv(key string) error {
	uid, err := d.GetEnvUID(key)
	if err != nil {
		return err
	}
	path := d.GetEnvPath(uid)
	err = d.Filehandler.DeleteFile(path)
	if err != nil {
		return err
	}
	return d.UpdateIndex(uid, "", key)
}

func (d *DynamicEnv) ListEnvs() ([]string, error) {
	index := d.LoadIndex()
	keys := make([]string, 0, len(*index))
	for _, iId := range *index {
		keys = append(keys, iId)
	}
	return keys, nil
}

func (d *DynamicEnv) ParseEnv(key string) (*DynamicEnvParsed, error) {
	parsed, err := d.GetEnv(key)
	if err != nil {
		return nil, errors.New("data not found: " + key)
	}

	result := DynamicEnvParsed{Local: make(map[string]string), Env: make(map[string]string)}

	if extends, ok := parsed.Data["extends"].([]any); ok {
		for _, dep := range extends {
			depKey, ok := dep.(string)
			if !ok {
				continue
			}
			parsedDep, err := d.ParseEnv(depKey)
			if err != nil {
				return nil, err
			}
			for k, v := range parsedDep.Local {
				result.Local[k] = v
			}
			for k, v := range parsedDep.Env {
				result.Env[k] = v
			}
		}
	}

	if local, ok := parsed.Data["local"].(map[string]any); ok {
		for k, v := range local {
			result.Local[k] = fmt.Sprintf("%v", v)
		}
	}

	if env, ok := parsed.Data["env"].(map[string]any); ok {
		for k, v := range env {
			value := fmt.Sprintf("%v", v)
			resolved := resolveEnvVariables(value, result.Local)
			result.Env[k] = resolved
		}
	}

	return &result, nil
}

func resolveEnvVariables(value string, local map[string]string) string {
	return os.Expand(value, func(variable string) string {
		if variable == "$" {
			return "$"
		}
		if val, ok := local[variable]; ok {
			return val
		}
		return ""
	})
}

func (d *DynamicEnv) GetEnvs(keys []string) map[string]string {
	envs := make(map[string]string)
	for _, key := range keys {
		parsed, err := d.ParseEnv(key)
		if err != nil {
			if d.Config.Debug {
				log.Printf("Error parsing env %s: %v\n", key, err)
			}
			continue
		}
		for k, v := range parsed.Env {
			envs[k] = v
		}
	}
	return envs
}

func (d *DynamicEnv) VerifyIdentities() error {
	cmd := exec.Command("age-keygen", "-y", d.Config.Identities)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to verify identities: %w", err)
	}

	identities := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, identity := range identities {
		for _, recipient := range d.UserConfig.Data.Recipients {
			if identity == recipient {
				return nil
			}
		}
	}

	return errors.New("no matching identity found in recipients")
}

func (d *DynamicEnv) ReencryptAll() error {
	err := d.VerifyIdentities()
	if err != nil {
		return err
	}

	envs := d.ListItems("")
	for _, value := range envs {
		d.SetEnv(value.Metadata.ID, value)
	}
	return nil
}

func (d *DynamicEnv) ExportTree(outDir string, prefix string) ([]string, error) {
	fs := filehandler.NewFileHandler(outDir, d.Config.Debug)
	envs := d.ListItems(prefix)
	fmt.Println("Loaded", len(envs), "files")
	keys := make([]string, len(envs))
	for _, value := range envs {
		key := value.Metadata.ID
		keys = append(keys, key)
		path, err := filepath.Rel(prefix, key)
		path = strings.ReplaceAll(path, "\\", "/")
		if err != nil || strings.HasPrefix(path, "..") {
			return nil, fmt.Errorf("failed to get relative path: %w", err)
		}

		output, err := d.FormatValue(value, false)
		if err != nil {
			return nil, fmt.Errorf("failed to format value: %w", err)
		}

		err = fs.WriteFile(path, output)
		if err != nil {
			return nil, fmt.Errorf("failed to write file: %w", err)
		}
	}
	return keys, nil
}

func (d *DynamicEnv) ImportTree(inDir string, prefix string) ([]string, error) {
	fs := filehandler.NewFileHandler(inDir, d.Config.Debug)
	files, err := fs.ListFiles("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	fmt.Println("Loaded", len(files), "files")
	keys := make([]string, len(files))
	for _, file := range files {
		value, err := fs.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		dynamicEnvValue, err := d.ParseRawValue(value, false)
		if err != nil {
			return nil, fmt.Errorf("failed to parse file: %w", err)
		}

		err = d.SetEnv(file, dynamicEnvValue)
		if err != nil {
			return nil, fmt.Errorf("failed to set env: %w", err)
		}
		keys = append(keys, file)
	}
	return keys, nil
}
