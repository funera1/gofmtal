package format

import (
	"bytes"
	"go/ast"
	"go/doc/comment"
	"go/format"
	"strings"

	"github.com/funera1/gofmtal/internal/derror"
)

/* TODO:
gofmtでprocessFileという名前がつけられてたから同じ名前つけていたが、
関数名から意味を読み取りにくいので、renameしても良さそう
*/
// 後で整理するためにprocessFileというFormatCodeの仮の関数の用意
func Format(filename string) (_ string, rerr error) {
	defer derror.Wrap(&rerr, "Format(%q)", filename)

	file, err := Parse(filename)
	if err != nil {
		return "", err
	}

	// 与えられたファイルからコメントを抜き出してすべてにフォーマットをかけて戻す
	for i, cmnts := range file.Syntax.Comments {
		for j, cmnt := range cmnts.List {
			formattedComment, err := formatCodeInComment(cmnt, file)
			if err != nil {
				return "", err
			}

			/*
				linesの調整はうまく行かないので一旦コメントアウト
				// フォーマットしたコメントをもとに戻す
				newlines, err := updateComment(cmnt, formattedComment, file)
				if err != nil {
					// コメントは行頭にあるという仮定で行の調整を行っているので、
					// コメントが行頭にない場合は調整を行わない
					continue
				}
				file.Lines = newlines
			*/
			cmnt.Text = formattedComment
			cmnts.List[j] = cmnt
		}

		file.Syntax.Comments[i] = cmnts
	}

	// formatでずれたlinesを調整する
	// file.Tfile.SetLines(file.Lines)

	var buf bytes.Buffer
	err = format.Node(&buf, file.Fset, file.Syntax)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// FormatCodeInComment はコメントを与えて、フォーマットしたコメントを返す
func formatCodeInComment(cmnt *ast.Comment, file *File) (_ string, rerr error) {
	defer derror.Debug(&rerr, "formatCodeInComment(%q)", cmnt.Text)

	commentString := cmnt.Text

	var p comment.Parser
	// p.Parseにつっこむときはコメントマーカー(//, /*, */)削除してから突っ込まないとだめ
	commentInfo := trimCommentMarker(commentString)
	doc := p.Parse(commentInfo.Comment)

	// commentStringからCodeを抜き出しその部分にだけフォーマットかける
	for _, c := range doc.Content {
		switch c := c.(type) {
		case *comment.Code:
			src, err := format.Source([]byte(c.Text))
			if err != nil {
				// format.Source()でsyntax errorが発生するコードは
				// そもそもformatできないので、無視する
				continue
			}

			c.Text = string(src)
		}
	}

	var pr comment.Printer
	b := pr.Comment(doc)
	formattedComment := string(b)

	// 改行するとコメントがずれるので削除
	formattedComment = strings.Trim(formattedComment, "\n")

	// コメントマーカーをつけ直す
	if commentInfo.CommentMarker == "//" {
		formattedComment = "// " + formattedComment
	} else {
		if commentInfo.LineCount == 1 {
			formattedComment = "/*" + formattedComment + "*/"
		} else {
			formattedComment = "/*\n" + formattedComment + "\n*/"
		}
	}

	return formattedComment, nil
}

/*
TrimCommentMarker はコメントからコメントマーカ(// や　/*)を取り除く
pkg.go.dev/go/doc/comment によると(commemt.Parser).Parseの引数にコメントを与えるとき
コメントマーカを削除してから与えることになっているため
*/

type CommentInfo struct {
	Comment       string
	CommentMarker string
	LineCount     int
}

func trimCommentMarker(comment string) CommentInfo {
	// 行数数える
	lineCount := strings.Count(comment, "\n")

	var commentMarker string

	// commentからcommentMarkerを取り除く
	if strings.HasPrefix(comment, "//") {
		commentMarker = "//"

		comment = strings.TrimLeft(comment, "//")
	} else {
		commentMarker = "/*"

		comment = strings.TrimLeft(comment, "/*")
		comment = strings.TrimRight(comment, "*/")
	}

	comment = strings.TrimLeft(comment, "\t")
	return CommentInfo{
		Comment:       comment,
		CommentMarker: commentMarker,
		LineCount:     lineCount,
	}
}
