package main

import "encoding/json"
import "log"
import "os"

type Config struct {
	data     map[string]interface{}
	filename string
}

func NewConfig(filename string) *Config {
	result := new(Config)
    result.filename = filename
	result.data = make(map[string]interface{})
	return result
}

// Saves a *Config into its marshaled json
func SaveConfig(config *Config) error {
    file, err := os.OpenFile(config.filename, os.O_CREATE, os.FileMode(0644))
    defer file.Close()
	if err != nil { return err }
    
    jsondata, err := json.MarshalIndent(config.data, "", "    ")
	if err != nil { return err }
    
    file.Write(jsondata)
    return nil
}

// Loads config information from a JSON file
func LoadConfig(filename string) *Config {
	result := NewConfig(filename)
	result.filename = filename
	err := result.parse()
	if err != nil {
		log.Fatalf("error loading config file %s: %s", filename, err)
	}
	return result
}

// Loads config information from a JSON string
func LoadConfigString(s string) *Config {
	result := NewConfig("")
	err := json.Unmarshal([]byte(s), &result.data)
	if err != nil {
		log.Fatalf("error parsing config string %s: %s", s, err)
	}
	return result
}

func (c *Config) parse() error {
	f, err := os.Open(c.filename)
	if err != nil {
		return err
	}
	defer f.Close()
	configdecode := json.NewDecoder(f)
    err = configdecode.Decode(&c.data)
	if err != nil {
		return err
	}

	return nil
}

// Set an object to the config
func (c *Config) Set(key string, thing interface{}) {
    c.data[key] = thing
}

// Returns a string for the config variable key
func (c *Config) GetString(key string) string {
	result, present := c.data[key]
	if !present {
		return ""
	}
	return result.(string)
}

// Returns an int for the config variable key
func (c *Config) GetInt(key string) int {
	x, ok := c.data[key]
	if !ok {
		return -1
	}
	return int(x.(float64))
}

// Returns a float for the config variable key
func (c *Config) GetFloat(key string) float64 {
	x, ok := c.data[key]
	if !ok {
		return -1
	}
	return x.(float64)
}

// Returns a bool for the config variable key
func (c *Config) GetBool(key string) bool {
	x, ok := c.data[key]
	if !ok {
		return false
	}
	return x.(bool)
}

// Returns an array for the config variable key
func (c *Config) GetArray(key string) []interface{} {
	result, present := c.data[key]
	if !present {
		return []interface{}(nil)
	}
	return result.([]interface{})
}