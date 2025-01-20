import base64

asset_go_out="cmd/httpd_asset.go"
template_dir="cmd/template"

rn="\r\n\r\n"
pkg="package cmd"


def fbase64(fpath: str = "") -> str:
	with open(fpath,"rb") as f:
		s=str(base64.b64encode(f.read()),'utf-8')
		f.close()
	return s

#
b64_video_js_min_css=fbase64(template_dir+"/video-js.min.css")
videojsmincss="".join([pkg,rn,"var videojsmincss string=`",b64_video_js_min_css,"`"])
#
b64_video_min_js=fbase64(template_dir+"/video.min.js")
videominjs="".join([rn,"var videominjs string=`",b64_video_min_js,"`"])
#
b64_style_css=fbase64(template_dir+"/style.css")
stylecss="".join([rn,"var stylecss string=`",b64_style_css,"`"])
#
b64_favicon_png=fbase64(template_dir+"/favicon.png")
faviconpng="".join([rn,"var faviconpng string=`",b64_favicon_png,"`"])
#
b64_video_bg_png=fbase64(template_dir+"/video-bg.png")
videobgpng="".join([rn,"var videobgpng string=`",b64_video_bg_png,"`"])

with open(asset_go_out,"w") as f:
	f.write(videojsmincss)
	f.write(videominjs)
	f.write(stylecss)
	f.write(faviconpng)
	f.write(videobgpng)
	f.close()
