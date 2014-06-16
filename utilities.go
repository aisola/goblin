package main

import "os"

// Check if File / Directory Exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

func IfTrueExec(thing bool, fn func(string, ...interface{}), format string, args ...interface{}) {
    if thing == true {
        fn(format, args...)
    }
}

func CreateSimpleFile(name, contents string, mode uint32) error {
    file, err := os.OpenFile(name, os.O_CREATE, os.FileMode(mode))
    defer file.Close()
	if err != nil { return err }
    
    file.WriteString(contents)
    return nil
}
