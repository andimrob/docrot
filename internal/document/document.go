package document

import (
	"bufio"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Document struct {
	Path      string
	Freshness *Freshness
	Content   string
}

type Freshness struct {
	LastReviewed string   `yaml:"last_reviewed"`
	Strategy     string   `yaml:"strategy"`
	Interval     string   `yaml:"interval,omitempty"`
	Expires      string   `yaml:"expires,omitempty"`
	Watch        []string `yaml:"watch,omitempty"`
	Ignore       []string `yaml:"ignore,omitempty"`
}

type frontmatterWrapper struct {
	Freshness *Freshness `yaml:"docrot"`
}

func Parse(path string) (*Document, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var lines []string
	inFrontmatter := false
	frontmatterLines := []string{}
	foundFrontmatter := false

	for scanner.Scan() {
		line := scanner.Text()

		if !foundFrontmatter && line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			} else {
				inFrontmatter = false
				foundFrontmatter = true
				continue
			}
		}

		if inFrontmatter {
			frontmatterLines = append(frontmatterLines, line)
		} else if foundFrontmatter {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	doc := &Document{
		Path:    path,
		Content: strings.Join(lines, "\n"),
	}

	if len(frontmatterLines) > 0 {
		var wrapper frontmatterWrapper
		yamlContent := strings.Join(frontmatterLines, "\n")
		if err := yaml.Unmarshal([]byte(yamlContent), &wrapper); err != nil {
			return nil, err
		}
		doc.Freshness = wrapper.Freshness
	}

	return doc, nil
}
