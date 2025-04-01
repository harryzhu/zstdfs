import os
import requests
import json
import random
import time
import mimetypes
from requests_toolbelt import MultipartEncoder
from blake3 import blake3
import pickle

size_limit = 16 << 20
print(f'max_upload_size: {size_limit}')

url_schema = "http://192.168.0.100:9090/api/upload/schema.json"
url_batch_import = "http://192.168.0.100:9090/api/batch-import"
url_upload = "http://192.168.0.100:9090/api/upload"
url_has = "http://192.168.0.100:9090/api/has"

user_apikey={}
user_apikey['harry'] = '16045879203066065922'

user_keys = user_apikey.keys()
item_schema = None

def load_schema():
	global item_schema	
	resp = requests.get(url_schema, timeout=5)
	print(resp.status_code)
	#print(resp.json())
	if resp.status_code == 200:
		item_schema = resp.json()
	#print(item_schema)

def upload_has(fpaths):
	data = {
		"fsums":[]
	}
	lines = []
	is_ok = False
	while 1:
		if is_ok:
			break
			return True
		try:
			for fpath in fpaths:
				with open(fpath,"rb") as f:
					fsum = blake3(f.read()).hexdigest()
					f.close()
					line = {}
					lines.append({fsum: fpath})
			#print(lines)
			data["fsums"] = json.dumps(lines)
			mdata = MultipartEncoder(fields=data)
			resp = requests.post(url_has, data=mdata, headers={'Content-Type': mdata.content_type}, timeout=15)
			print(resp.status_code) 
			print(resp.text) 
			if resp.status_code == 200:
				is_ok = True
				return resp.json()
		except Exception as err:
			print(err)
			time.sleep(5)
			


def upload_file(fpath, username):
	if fpath=="" or username == "":
		return None
	if username not in user_keys:
		return None
	if user_apikey[username] == "":
		return None
	#
	data = {
		'fuser': username,
		'fapikey': user_apikey[username],
	}
	try:
		f1 = open(fpath, 'rb')
		files = {'file': f1}
		resp = requests.post(url_batch_import, files=files, data=data, timeout=3)
		print(resp.status_code) 
		print(resp.text) 
		f1.close()
	except Exception as err:
		print(err)

multi_files = {}
idx = 0
sess = requests.session()
sess.keep_alive = False
sess.mount('https://', requests.adapters.HTTPAdapter(pool_connections=100, pool_maxsize=200))
sess.mount('http://', requests.adapters.HTTPAdapter(pool_connections=100, pool_maxsize=200))

def upload_batch(fpaths, username):
	if len(fpaths)==0 or username == "":
		return None
	if username not in user_apikey.keys():
		return None
	if user_apikey[username] == "":
		return None
	#
	data = {
		'fuser': username,
		'fapikey': user_apikey[username],
	}
	global multi_files
	global idx
	#global total_size
	idx = 0
	total_size = 0
	for fpath in fpaths:		
		k = f'k{idx}'
		#print(f'{k}: {fpath}')
		total_size += os.path.getsize(fpath)
		multi_files[k]=(os.path.basename(fpath), open(fpath, 'rb'), "video/mp4")
		idx += 1

	is_ok = False
	while 1:
		if is_ok:
			break
			return True

		try:
			print(len(multi_files))
			print("============")
			headers = {'Connection': 'close',}
			resp = sess.post(url_batch_import, files=multi_files, data=data, headers=headers, timeout=15)
			print(resp.status_code) 
			#print(resp.text)
			resp.close()
			is_ok = True
		except Exception as err:
			print(err)
			is_ok = False
			print(f'total_size: {round(total_size/1024/1024)} MB')
			print("waiting ...")
			time.sleep(5)
		finally:
			pass
			#print(f'total_size: {round(total_size/1024/1024)} MB')

	return True


def upload_meta(fpath, username):
	if fpath=="" or username == "":
		return None
	if username not in user_apikey.keys():
		return None
	if user_apikey[username] == "":
		return None
	#
	data = item_schema
	data["fuser"] = username
	data["fapikey"] = user_apikey[username]
	data["file"] = None
	#print(data)
	try:
		with open(fpath,"rb") as f:
			fsum = blake3(f.read()).hexdigest()
			f.close()
			paths = fpath.split("/")
			fid = "/".join(paths[1:])
			print(f"fid:{fid}")
			finfo = os.stat(fpath)
			meta = {
			"size": str(round(finfo.st_size)),
			"mtime": str(round(finfo.st_mtime)),
			"mime": "video/mp4",
			"fsum": fsum,
			"is_public": "1",
			"is_ban": "0"
			}
			
			data["fid"] = fid
			data["fmeta"] = json.dumps(meta)			
			#print(mdata)			
			if is_with_file:
				files = {'file': open(fpath, 'rb')}
				resp = requests.post(url_upload, files=files, data=data)
			else:
				mdata = MultipartEncoder(fields=data)
				resp = requests.post(url_upload, data=mdata, headers={'Content-Type': mdata.content_type})
			print(resp.status_code) 
			print(resp.text) 
	except Exception as err:
		print(err)

def batch_import(dpath, meta_or_file):
	batch = []
	total_size = 0
	for root, dirs, files in os.walk(dpath, True):
		for f in files:
			if f[-4:].lower() != ".mp4":
				continue
			fpath = os.path.join(root,f).replace("\\","/")
			if os.path.getsize(fpath) > size_limit:
				continue
			#print(fpath)
			if meta_or_file == "meta":
				upload_meta(fpath,"harry")

			if meta_or_file == "file":
				upload_file(fpath, 'harry')

			if meta_or_file == "batch":
				if len(batch) < 49:
					batch.append(fpath)
				else:
					batch.append(fpath)
					#print("batch.upload",batch)
					print("batch.upload len:",len(batch))
					t3 = time.time()
					has_files = upload_has(batch)
					print(has_files)
					for kv in has_files:
						for k,v in kv.items():
							if v == "1":
								batch.remove(k)
					print("----filtered batch------")
					#print(batch)
					if len(batch) > 0:
						upload_batch(batch, 'harry')
					t4 = time.time()
					print(f'Running Time: {t4 - t3} seconds, uploaded files: {len(batch)}')
					batch.clear()
					print("batch.clear",batch)
					#break
			

			

# t1 = time.time()
# is_with_file = True
# load_schema()
# batch_import("v_best", "meta")
# t2 = time.time()


# t1 = time.time()
# is_with_file = False
# load_schema()
# batch_import("v_best", "file")
# # is_with_file = False
# #batch_import("v_best", "meta")
# t2 = time.time()

# t1 = time.time()
# load_schema()
# batch_import("v_best", "file")
# #batch_import("v_better", "batch")
# t2 = time.time()

t1 = time.time()
load_schema()
#batch_import("v_best", "batch")
batch_import("v_better", "batch")
t2 = time.time()

print(f"Elapse: {t2-t1} sec")
#batch_import("v_better")
#batch_import("v")


