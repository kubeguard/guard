package token

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/pkg/errors"
	authv1 "k8s.io/api/authentication/v1"
)

type Authenticator struct {
	options  Options
	tokenMap map[string]authv1.UserInfo
	lock     sync.RWMutex
}

func New(opts Options) *Authenticator {
	return &Authenticator{
		options:  opts,
		tokenMap: map[string]authv1.UserInfo{},
	}
}

func (s *Authenticator) Configure() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	data, err := LoadTokenFile(s.options.AuthFile)
	if err != nil {
		return err
	}
	s.tokenMap = data
	return nil
}

func (s *Authenticator) Check(token string) (*authv1.UserInfo, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	user, ok := s.tokenMap[token]
	if !ok {
		return nil, errors.New("Invalid token")
	}
	return &user, nil
}

//https://kubernetes.io/docs/admin/authentication/#static-token-file
//csv token file:
//  - four field required (format : token,user,uid,"group1,group2,group3")
//  - groups can be empty, others cannot be empty
//  - token should be unique
//  - one user can have multiple token
func LoadTokenFile(file string) (map[string]authv1.UserInfo, error) {
	csvFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer csvFile.Close()

	reader := csv.NewReader(bufio.NewReader(csvFile))
	reader.FieldsPerRecord = -1
	data := map[string]authv1.UserInfo{}
	lineNum := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, errors.Wrap(err, "failed to parse token auth file")
		}
		lineNum++
		cols := len(row)

		if cols < 3 || cols > 4 {
			return nil, errors.Errorf("line #%d of token auth file is ill formatted", lineNum)
		}

		token := strings.TrimSpace(row[0])
		if len(token) == 0 {
			return nil, errors.Errorf("line #%d of token auth file has empty token", lineNum)
		}
		if _, found := data[token]; found {
			return nil, errors.Errorf("line #%d of token auth file reuses token", lineNum)
		}

		user := authv1.UserInfo{
			Username: strings.TrimSpace(row[1]),
			UID:      strings.TrimSpace(row[2]),
		}
		if user.Username == "" {
			return nil, errors.Errorf("line #%d of token auth file has empty user name", lineNum)
		}
		if user.UID == "" {
			return nil, errors.Errorf("line #%d of token auth file has empty uid", lineNum)
		}

		if cols > 3 {
			user.Groups = parseGroups(strings.TrimSpace(row[3]))
		}
		data[token] = user
	}
	return data, nil
}

//string format : "group1,group2,group3"
func parseGroups(in string) []string {
	var out []string
	groups := strings.Split(in, ",")
	for _, g := range groups {
		if len(g) > 0 {
			out = append(out, strings.TrimSpace(g))
		}
	}
	return out
}
