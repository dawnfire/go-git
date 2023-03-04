package main

import (
	"errors"
	"fmt"

	"os"
	"path/filepath"
	"regexp"

	"github.com/go-git/go-git/v5"
	. "github.com/go-git/go-git/v5/_examples"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
)

type Msg struct {
	Data []byte
	Msg  string
}

type Q struct {
	Err error
	Msg
}

func (q *Q) RespErr() {
	fmt.Println(q.Err.Error())
}

func (q *Q) Resp() {
	fmt.Println(q.Msg.Msg)
}

type zLog struct {
}

func (z *zLog) Error(s string) {
	fmt.Println(s)
}

var z = &zLog{}

var (
	repoArena = "/tmp/repositories"
	user      = "kzz"
	token     = "118200"
	repoUrl   = "https://localhost/git/kzz/gittravel.git"
)

func openRepoInMemory() (repo *git.Repository, err error) {
	q := &Q{}
	pattern := regexp.MustCompile("(?i)^http(s)?://(?P<name>.*/.*)\\.git$")
	match := pattern.FindStringSubmatch(repoUrl)

	if len(match) <= 0 {
		q.Err = fmt.Errorf("must provide valid repo")
		q.RespErr()
		err = q.Err
		return
	}
	matchResult := make(map[string]string)
	for k, v := range pattern.SubexpNames() {
		if k == 0 || v == "" {
			continue
		}
		matchResult[v] = match[k]
	}
	repoName := matchResult["name"]
	if repoName == "" {
		q.Err = fmt.Errorf("invalid repo name")
		q.RespErr()
		err = q.Err
		return
	}

	cloneOption := &git.CloneOptions{
		URL:          repoUrl,
		SingleBranch: false,
	}

	auth := http.BasicAuth{
		Username: user,
		Password: token,
	}

	cloneOption.Auth = &auth

	repo, q.Err = git.Clone(memory.NewStorage(), nil, cloneOption)
	if q.Err != nil {
		q.RespErr()
		err = q.Err
		return
	}

	if q.Err != nil {
		q.RespErr()
		err = q.Err
		return
	}

	if repo == nil {
		q.Err = fmt.Errorf("repo is nil")
		q.RespErr()
		err = q.Err
	}

	return
}

func openRepo() (repo *git.Repository, err error) {
	q := &Q{}

	pattern := regexp.MustCompile("(?i)^http(s)?://(?P<name>.*/.*)\\.git$")
	match := pattern.FindStringSubmatch(repoUrl)

	if len(match) <= 0 {
		q.Err = fmt.Errorf("must provide valid repo")
		q.RespErr()
		err = q.Err
		return
	}
	matchResult := make(map[string]string)
	for k, v := range pattern.SubexpNames() {
		if k == 0 || v == "" {
			continue
		}
		matchResult[v] = match[k]
	}
	repoName := matchResult["name"]
	if repoName == "" {
		q.Err = fmt.Errorf("invalid repo name")
		q.RespErr()
		err = q.Err
		return
	}

	cloneOption := &git.CloneOptions{
		URL:          repoUrl,
		SingleBranch: false,
	}

	auth := http.BasicAuth{
		Username: user,
		Password: token,
	}

	cloneOption.Auth = &auth

	repoDir := filepath.Join(repoArena, repoName)

	fileInfo, err := os.Stat(repoDir)
	q.Err = err
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		q.RespErr()
		return
	}

	if err == nil && !fileInfo.IsDir() {
		q.Err = fmt.Errorf("%s exists in %s as a file", repoName, repoArena)
		q.RespErr()
		return
	}

	q.Err = nil

	if errors.Is(err, os.ErrNotExist) {
		repo, q.Err = git.PlainClone(repoDir, false, cloneOption)
		if q.Err != nil {
			q.RespErr()
			err = q.Err
			return
		}

	} else {
		repo, q.Err = git.PlainOpen(repoDir)
		if q.Err != nil {
			q.RespErr()
			err = q.Err
			return
		}
	}
	err = nil

	if repo == nil {
		q.Err = fmt.Errorf("repo is nil")
		q.RespErr()
		err = q.Err
	}

	return
}

func refLog(repo *git.Repository, action string,
	parent, current *plumbing.Reference, from string) (err error) {
	if repo == nil || parent == nil || current == nil || action == "" {
		err = fmt.Errorf("empty/nil param(s): repo|action|parent")
		z.Error(err.Error())
		return
	}
	//0000000000000000000000000000000000000000
	//f467e2d133bc9b393c1f9548d178e547b63ff76f
	//dawnfire <dawnfire@126.com> 1675246377 +0800
	//clone:
	//from
	//https://localhost/git/kzz/gittravel.git

	cfg, err := repo.ConfigScoped(config.GlobalScope)
	if err != nil {
		z.Error(err.Error())
		return
	}

	pattern := map[string]string{
		"clone": "%s %s %s\tclone: from %s",
	}
	switch action {
	case "clone":
		operator := fmt.Sprintf("%s %s", cfg.User.Name, cfg.User.Email)
		msg := plumbing.ReferenceName(fmt.Sprintf(pattern[action],
			parent.Hash().String(),
			current.Hash().String(),
			operator, from))
		name := plumbing.NewLogReferenceName(current.Name().String())
		logRef := plumbing.NewLogReference(name, msg)
		err = repo.Storer.SetLog(logRef)
		if err != nil {
			z.Error(err.Error())
			return
		}
	}

	return
}

func main() {
	repo, err := openRepo()
	//repo, err := openRepoInMemory()
	if err != nil {
		return
	}

	// Gets the HEAD history from HEAD, just like this command:
	Info("git log")

	//msg := plumbing.ReferenceName("x5--f467e2d133bc9b393c1f9548d178zzzze547b63ff76f 06026f047be9fabb9001e52a221b70ac5dafebe2 dawnfire <dawnfire@126.com> 1675214814 +0800\tcheckout: moving from master to b1")
	//branchName := "refs/heads/s1/s2/s3"
	////branchName := "HEAD"
	//
	//name := plumbing.NewLogReferenceName(branchName)
	//
	//// setup .git/HEAD with content refs/heads/<branchName>
	//logRef := plumbing.NewLogReference(name, msg)
	//err = repo.Storer.SetLog(logRef)
	//if err != nil {
	//	z.Error(err.Error())
	//	return
	//}
	cfg, err := repo.Config()
	remoteCfg := cfg.Remotes["origin"]

	head, err := repo.Head()
	//fmt.Println(remoteCfg.URLs[0], head.Hash())
	err = refLog(repo, "clone",
		plumbing.NewHashReference(plumbing.ReferenceName(""), plumbing.Hash{}),
		head, remoteCfg.URLs[0],
	)

	refName := plumbing.ReferenceName("logs/refs/heads/master")
	r, err := repo.RefLog(refName, true)
	if err != nil || r == nil {
		fmt.Println(err.Error())
		return
	}

	r, err = repo.RefLog(refName, false)
	if err != nil || r == nil {
		fmt.Println(err.Error())
		return
	}
}
