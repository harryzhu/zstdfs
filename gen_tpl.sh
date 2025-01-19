rm -f template/.DS_Store
python asset2go.py
go-bindata -o cmd/httpd_tpl.go -pkg cmd -ignore=.gitignore template/