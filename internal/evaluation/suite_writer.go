package evaluation

import (
	"bufio"
	"encoding/json"
	"os"
)

// WriteSuiteJSONL writes suite records atomically to path.
func WriteSuiteJSONL(path string, records []SuiteRecord) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(f)
	for _, rec := range records {
		line, err := json.Marshal(rec)
		if err != nil {
			f.Close()
			os.Remove(tmp)
			return err
		}
		if _, err := writer.Write(append(line, '\n')); err != nil {
			f.Close()
			os.Remove(tmp)
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}
