package envfile

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Parse will load one or many env files into an Env object
func Parse(env map[string]string, paths ...string) error {
	if env == nil {
		return errors.New("cannot add values to nil map")
	}
	for _, path := range paths {
		if err := loadEnvFile(path, env); err != nil {
			return err
		}
	}
	return nil
}

func loadEnvFile(filename string, env map[string]string) error {
	if filename == "" {
		return nil
	}
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return parseEnvFile(f, env)
}

func parseEnvFile(f io.Reader, env map[string]string) error {
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		line = os.Expand(line, func(v string) string {
			if val, ok := env[v]; ok {
				return val
			}
			return os.Getenv(v)
		})
		if eq := strings.Index(line, "="); eq >= 0 && len(line) > eq {
			key := strings.TrimSpace(line[:eq])
			val := strings.TrimSpace(line[eq+1:])
			if strings.HasPrefix(line[eq+1:], "<<") {
				heredoc := strings.Split(val, " ")[0]
				env[key] = parseBlock(val, heredoc, strings.TrimPrefix(heredoc, "<<"), scanner)
			} else if strings.HasPrefix(val, `"""`) {
				env[key] = parseBlock(val, `"""`, `"""`, scanner)
			} else if strings.HasPrefix(val, "'''") {
				env[key] = parseBlock(val, "'''", "'''", scanner)
			} else {
				env[key] = strings.Trim(val, "\"'")
			}
		}
	}
	return scanner.Err()
}

func parseBlock(value, start, end string, scanner *bufio.Scanner) string {
	parts := []string{}
	firstLine := strings.TrimPrefix(value, start)
	fmt.Println(value, firstLine, start, start == value)
	if strings.TrimSpace(firstLine) != "" {
		parts = append(parts, strings.TrimSpace(firstLine))
	}
	for scanner.Scan() {
		if part := scanner.Text(); part == end {
			break
		} else {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, "\n")
}
