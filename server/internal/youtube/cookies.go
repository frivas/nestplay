package youtube

import (
	"strings"
)

func cookieHeader(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if !strings.Contains(raw, "\t") {
		return collapseHeader(raw)
	}
	type pair struct {
		name  string
		value string
		rank  int
	}
	var pairs []pair
	index := map[string]int{}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 7 {
			continue
		}
		domain := strings.ToLower(fields[0])
		if !strings.Contains(domain, "youtube.com") && !strings.Contains(domain, "google.com") {
			continue
		}
		name := strings.TrimSpace(fields[5])
		value := strings.TrimSpace(fields[6])
		if name == "" || value == "" {
			continue
		}
		rank := 1
		if strings.Contains(domain, "youtube.com") || strings.Contains(domain, "youtu.be") {
			rank = 0
		}
		item := pair{name: name, value: value, rank: rank}
		if i, ok := index[name]; ok {
			if rank <= pairs[i].rank {
				pairs[i] = item
			}
			continue
		}
		index[name] = len(pairs)
		pairs = append(pairs, item)
	}
	parts := make([]string, 0, len(pairs))
	for _, p := range pairs {
		parts = append(parts, p.name+"="+p.value)
	}
	return strings.Join(parts, "; ")
}

func collapseHeader(raw string) string {
	index := map[string]int{}
	var parts []string
	for _, chunk := range strings.FieldsFunc(raw, func(r rune) bool {
		return r == '\n' || r == '\r' || r == ';'
	}) {
		chunk = strings.TrimSpace(chunk)
		if chunk == "" {
			continue
		}
		name, value, ok := strings.Cut(chunk, "=")
		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)
		if !ok || name == "" || value == "" {
			continue
		}
		item := name + "=" + value
		if i, ok := index[name]; ok {
			parts[i] = item
			continue
		}
		index[name] = len(parts)
		parts = append(parts, item)
	}
	return strings.Join(parts, "; ")
}
