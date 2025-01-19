import base64

asset_go_out="cmd/httpd_asset.go"
template_dir="template"

rn="\r\n\r\n"
pkg="package cmd"
css=""
js=""
favicon=""
stylecss=""

# css
video_js_min_css=""
with open(template_dir+"/video-js.min.css","r") as f:
	video_js_min_css=str(base64.b64encode(f.read().encode('utf-8')),'utf-8')
	f.close()
css="".join([pkg,rn,"var videojsmincss string=`",video_js_min_css,"`"])

# js
video_min_js=""
with open(template_dir+"/video.min.js","r") as f:
	video_min_js=str(base64.b64encode(f.read().encode('utf-8')),'utf-8')
	f.close()
js="".join([rn,"var videominjs string=`",video_min_js,"`"])

# style.css
style_css=""
with open(template_dir+"/style.css","r") as f:
	style_css=str(base64.b64encode(f.read().encode('utf-8')),'utf-8')
	f.close()
stylecss="".join([rn,"var stylecss string=`",style_css,"`"])

# favicon
favicon_png=""
with open(template_dir+"/favicon.png","rb") as f:
	favicon_png=str(base64.b64encode(f.read()), 'utf-8')
	f.close()
favicon="".join([rn,"var faviconpng string=`",favicon_png,"`"])

# write go
with open(asset_go_out,"w") as f:
	f.write(css)
	f.write(stylecss)
	f.write(js)
	f.write(favicon)
	f.close()




