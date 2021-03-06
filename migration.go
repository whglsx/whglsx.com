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
	reDoc := regexp.MustCompile(`(.*)\.doc`)
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
			p.WriteContentDoc(f)
		}
		if p.Thumbnail != "" {
			copyToResource(filepath.Join(c.path, p.Thumbnail), dir)
		} else {
			fmt.Println("Can't find head!", c.Name, p.Name)
		}
		for _, img := range p.ContentImage {
			copyToResource(filepath.Join(c.path, img), dir)
		}
	}
}
func copyToResource(path, base string) {
	cmd := exec.Command("cp", path, filepath.Join(base, "static", "product_img"))
	opb, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("E:", err, string(opb))
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
Title: "%s"
categories:
    - "%s"
thumbnail: "%s"
contentImages: %s
---
`, p.Name, p.category, p.Thumbnail, c))
}

func (p *Product) WriteContentDoc(w io.Writer) {
	basename := filepath.Base(p.ContentDoc)
	basedir := filepath.Dir(p.ContentDoc)
	basename = basename[:len(basename)-4]
	fmt.Println("BaseDIR:", basedir)
	htmlFile := filepath.Join(basedir, basename+".html")
	c := exec.Command("soffice", "--headless", "--convert-to", "html", p.ContentDoc)
	os.Chdir(basedir)
	c.Start()
	c.Wait()
	cmd := exec.Command("pandoc", "-f", "html", "-t", "markdown_github", htmlFile)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("ER:", string(bytes))
		panic(err)
	}
	_, err = w.Write(bytes)
	if err != nil {
		panic(err)
	}
	exec.Command("rm", htmlFile)
}

func CopyResource() {
}
func main() {
	base := "/home/snyh/prj/whglsx/data"
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
