package format

import (
	"bytes"
	"go/doc/comment"
	"go/format"
	"path/filepath"
	"strings"
)

/*
TrimCommentMarker はコメントからコメントマーカ(// や　/*)を取り除く
pkg.go.dev/go/doc/commentによると(commemt.Parser).Parseの引数にコメントを与えるとき
コメントマーカを削除してから与えることになっているため
*/
func TrimCommentMarker(comment string) (string, string) {
	var commentMarker string
	if strings.HasPrefix(comment, "//") {
		comment = strings.TrimLeft(comment, "//")
		commentMarker = "//"
	} else {
		comment = strings.TrimLeft(comment, "/*")
		comment = strings.TrimRight(comment, "*/")
		commentMarker = "/*"
	}
	comment = strings.TrimLeft(comment, "\t")
	return comment, commentMarker
}

// FormatCodeInComment はコメントを与えて、フォーマットしたコメントを返す
func FormatCodeInComment(commentString string) (string, error) {
	var p comment.Parser
	// p.Parseにつっこむときはコメントマーカー(//, /*, */)削除してから突っ込まないとだめ
	c, commentMarker := TrimCommentMarker(commentString)
	doc := p.Parse(c)

	// commentStringからCodeを抜き出しその部分にだけフォーマットかける
	for _, c := range doc.Content {
		switch c := c.(type) {
		case *comment.Code:
			src, err := format.Source([]byte(c.Text))
			if err != nil {
				return "", err
			}
			c.Text = string(src)
		}
	}

	// コメントから抜き出したコードについてフォーマットをかける
	var pr comment.Printer
	b, err := format.Source(pr.Comment(doc))
	if err != nil {
		return "", err
	}
	formattedComment := string(b)

	// 改行するとコメントがずれるので削除
	formattedComment = strings.Trim(formattedComment, "\n")

	// コメントマーカーをつけ直す
	if commentMarker == "//" {
		formattedComment = "// " + c
	} else {
		formattedComment = "/*\n" + c + "\n*/"
	}

	return formattedComment, nil
}

// 後で整理するためにprocessFileというFormatCodeの仮の関数の用意
func processFile(filename string) (string, error) {
	// TODO: fはわかりにくそう
	astFile, fset, err := GetAst(filename)
	if err != nil {
		return "", err
	}

	// 与えられたファイルからコメントを抜き出してすべてにフォーマットをかけて戻す
	// cmnts: astからcommentGroupを抜き出したもの
	// cmnt: commentGroupからcommnetを抜き出したもの
	for i, cmnts := range astFile.Comments {
		for j, cmnt := range cmnts.List {
			formattedComment, err := FormatCodeInComment(cmnt.Text)
			if err != nil {
				return "", err
			}

			// フォーマットしたコメントをもとに戻す
			cmnt.Text = formattedComment
			cmnts.List[j] = cmnt
		}

		astFile.Comments[i] = cmnts
	}

	var buf bytes.Buffer
	err = format.Node(&buf, fset, astFile)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func IsGoFile(filename string) bool {
	return (filepath.Ext(filename) == ".go")
}
