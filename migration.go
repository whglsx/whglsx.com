package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
)

type Category struct {
	Name     string
	path     string
	Products map[string]*Product
}

func NewCategory(name, dirpath string) *Category {
	c := &Category{name, dirpath, make(map[string]*Product)}
	/*
		pname.html
		pname_head.(jpg|gif)
		pname_1.(jpg|gif)
	*/
	infos, err := ioutil.ReadDir(dirpath)
	if err != nil {
		panic(err)
	}
	reDoc := regexp.MustCompile(`(.*)\.html`)
	reThumbnail := regexp.MustCompile(`(.*)_head\.*`)
	reCImgs := regexp.MustCompile(`(.*)_(\d)+\..*`)
	for _, info := range infos {
		{
			r := reDoc.FindStringSubmatch(info.Name())
			if len(r) == 2 {
				c.getProduct(r[1]).ContentDoc = filepath.Join(c.path, info.Name())
			}
		}
		{
			r := reThumbnail.FindStringSubmatch(info.Name())
			if len(r) == 2 {
				c.getProduct(r[1]).Thumbnail = info.Name()
			}
		}
		{
			r := reCImgs.FindStringSubmatch(info.Name())
			if len(r) == 3 {
				//TODO: handle pos
				pos, err := strconv.Atoi(r[2])
				if err != nil {
					panic(err)
				}
				c.getProduct(r[1]).ContentImage[pos] = info.Name()
			}
		}

	}
	return c
}
func (c *Category) getProduct(name string) *Product {
	if p, ok := c.Products[name]; ok {
		return p
	} else {
		p := &Product{Name: name, ContentImage: make(map[int]string)}
		p.category = c.Name
		c.Products[name] = p
		return p
	}
}

func (c *Category) Gen(dir string) {
	for _, p := range c.Products {
		f, err := os.Create(filepath.Join(dir, "content", "product", p.Name+".md"))
		defer f.Close()
		if err != nil {
			panic(err)
		}
		p.WriteHeader(f)
		if p.ContentDoc != "" {
			p.WriteContentHtml(f)
		}
		if p.Thumbnail != "" {
			cmd := exec.Command("cp", filepath.Join(c.path, p.Thumbnail), filepath.Join(dir, "static", "product_img"))
			opb, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Println("E:", err, string(opb))
			}
		}
	}
}

type Product struct {
	Name         string
	Thumbnail    string
	ContentDoc   string
	ContentImage map[int]string
	category     string
}

func (p *Product) WriteHeader(w io.Writer) {
	//write header
	c := "["

	for i := 1; i < len(p.ContentImage)+1; i++ {
		if i != 1 {
			c += ","
		}
		c += `"` + p.ContentImage[i] + `"`
	}
	c += "]"
	io.WriteString(w, fmt.Sprintf(`---
Title: "%s",
Category:
   - "%s"
Thumbnail: "%s"
ContentImage: %s
---
`, p.Name, p.category, p.Thumbnail, c))
}

func (p *Product) WriteContentHtml(w io.Writer) {
	cmd := exec.Command("pandoc", "-f", "html", "-t", "markdown_github", p.ContentDoc)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("ER:", string(bytes))
		panic(err)
	}
	_, err = w.Write(bytes)
	if err != nil {
		panic(err)
	}
}

func CopyResource() {
}
func main() {
	base := "/dev/shm/x"
	infos, err := ioutil.ReadDir(base)

	if err != nil {
		panic(err)
	}
	for _, info := range infos {
		if info.IsDir() {
			c := NewCategory(info.Name(), filepath.Join(base, info.Name()))
			c.Gen("/home/snyh/prj/whglsx")
			fmt.Println(len(c.Products))
		}
	}

}
