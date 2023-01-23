package github

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type GithubDao struct {
}

func NewGithubDao() GithubDao {

	return GithubDao{}
}

func (dao GithubDao) GetDocumentList(url string) []string {
	buf, err := dao.GetHttpBodyBuf(url)
	if err != nil {
		return nil
	}
	return strings.Split(string(buf), "\n")
}

func (dao GithubDao) GetMarkdown(url string) ([]byte, error) {
	return dao.GetHttpBodyBuf(url)
}

func (dao GithubDao) GetTemplate(url string) ([]byte, error) {
	return dao.GetHttpBodyBuf(url)
}

func (dao GithubDao) GetHttpBodyBuf(url string) ([]byte, error) {
	log.Println("downloading file from: " + url)
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("get http file, %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status, %d", response.StatusCode)
	}

	buf, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response, %w", err)
	}

	return buf, err
}
