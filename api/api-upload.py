import os
import requests
import json
import random
import time
import mimetypes
import hashlib
from requests_toolbelt import MultipartEncoder
import pickle

size_limit = 16 << 20
print(f'max_upload_size: {size_limit}')

url_schema = "http://127.0.0.1:9090/api/upload/schema.json"
url_upload = "http://127.0.0.1:9090/api/upload"

user_apikey={}
user_apikey['harry'] = '14125811004486689209'

user_keys = user_apikey.keys()

sess = requests.session()
sess.keep_alive = False
sess.mount('https://', requests.adapters.HTTPAdapter(pool_connections=100, pool_maxsize=200))
sess.mount('http://', requests.adapters.HTTPAdapter(pool_connections=100, pool_maxsize=200))

global item_schema


def load_schema():
	global item_schema	
	resp = requests.get(url_schema, timeout=5)
	print(resp.status_code)
	if resp.status_code == 200:
		item_schema = resp.json()
		del item_schema['file']
		

def file_sha256(fpath):
	h = hashlib.sha256()
	with open(fpath, "rb") as f1:
		h.update(f1.read())
	return h.hexdigest()

def upload_file(fpath, username, dpath):
	if fpath=="" or username == "":
		return None
	if username not in user_keys:
		return None
	if user_apikey[username] == "":
		return None
	#
	finfo = os.stat(fpath)

	data = {}
	
	data["fuser"] = username
	data["fapikey"] = user_apikey[username]
	data["fid"] = fpath[len(dpath):].strip("/")
	if data["fid"][0:1] == ".":
		return 

	fmeta = {}
	fmeta["size"] = str(round(finfo.st_size))
	fmeta["mtime"] = str(round(finfo.st_mtime))
	fmeta["mime"] = mimetypes.guess_type(fpath)[0]
	fmeta["is_public"] = "1"
	fmeta["is_ban"] = "0"
	fmeta["tags"] = ""
	fmeta["stats_digg_count"] = "0"
	fmeta["stats_collect_count"] = "0"
	fmeta["stats_share_count"] = "0"
	fmeta["stats_comment_count"] = "0"
	fmeta["stats_download_count"] = "0"
	fmeta["dot_color"] = ""
	
	try:
		f1 = open(fpath, 'rb')
		fmeta["fsha256"] = file_sha256(fpath)
		#
		data["fmeta"] = json.dumps(fmeta, ensure_ascii=False)
		print("======")
		print(data)
		files = {'file': f1}
		resp = sess.post(url_upload, files=files, data=data, timeout=3)
		print(resp.status_code) 
		print(resp.text) 
		f1.close()
	except Exception as err:
		print(err)



def batch_import(dpath):
	rdir = dpath.replace("\\","/")
	num = 0
	for root, dirs, files in os.walk(dpath, True):
		for f in files:
			if f[-4:].lower() != ".mp4":
				continue
			fpath = os.path.join(root,f).replace("\\","/")
			if os.path.getsize(fpath) > size_limit:
				continue
			print(fpath)

			upload_file(fpath, 'harry', rdir)
			num += 1
			

			
#

t1 = time.time()
load_schema()
print("item_schema: ", item_schema)
root_dir = "/Users/harry/Desktop/v"
batch_import(root_dir)
t2 = time.time()

print(f"Elapse: {t2-t1} sec")



