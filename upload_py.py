import os
import requests

#
endpoint = "http://192.168.0.106:8080/admin/upload"
admin_user = "admin"
admin_password = "123"
#
import_dir = "/Users/harry/Downloads/bootstrap-5.3.3-dist/css"
import_ext = ".css"
#
import_user = "harry"
import_group = "bootstrap"
import_prefix = "v5.3.3/css"

for root, dirs, files in os.walk(import_dir):
	for f in files:
		fpath = os.path.join(root,f)
		fname, fext = os.path.splitext(f)
		if fext.lower() != import_ext:
			continue
		print("import:",fpath)
		files = {
			"file": open(fpath,"rb")
		}
		form_kv = {
			"fuser": import_user,
			"fgroup": import_group,
			"fprefix": import_prefix
		}

		response = requests.post(endpoint, data=form_kv, files=files,auth=(admin_user,admin_password))
		print(response.text)





