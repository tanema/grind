package envfile

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Env contains loaded envfiles for uses for command execution
type Env struct {
	vars map[string]string
}

// Parse will load one or many env files into an Env object
func Parse(paths ...string) (*Env, error) {
	env := &Env{vars: map[string]string{}}
	for _, path := range paths {
		if err := env.loadEnvFile(path); err != nil {
			return nil, err
		}
	}
	return env, nil
}

// Export will export the variables to the running application
func (env *Env) Export() error {
	for key, val := range env.vars {
		if err := os.Setenv(key, val); err != nil {
			return err
		}
	}
	return nil
}

// ToArray will export as an env array for use in exec.Cmd env.
func (env *Env) ToArray() []string {
	vars := []string{}
	for key, val := range env.vars {
		vars = append(vars, key+"="+val)
	}
	return vars
}

func (env *Env) loadEnvFile(filename string) error {
	if filename == "" {
		return nil
	}
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to Open %q: %v", filename, err)
	}
	defer f.Close()
	return env.parseEnvFile(f)
}

func (env *Env) parseEnvFile(f io.Reader) error {
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		if eq := strings.Index(line, "="); eq >= 0 && len(line) > eq {
			key := strings.TrimSpace(line[:eq])
			if strings.HasPrefix(line[eq+1:], "<<") {
				env.vars[key] = parseHeredoc(line[eq+1:], scanner)
			} else {
				env.vars[key] = strings.Trim(strings.TrimSpace(line[eq+1:]), "\"")
			}
		}
	}
	return scanner.Err()
}

func parseHeredoc(value string, scanner *bufio.Scanner) string {
	heredoc := strings.Split(value[2:], " ")[0]
	firstLine := strings.TrimPrefix(value, "<<"+heredoc)
	parts := []string{}
	if strings.TrimSpace(firstLine) != "" {
		parts = append(parts, strings.TrimSpace(firstLine))
	}
	for scanner.Scan() {
		if part := scanner.Text(); part == heredoc {
			break
		} else {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, "\n")
}
