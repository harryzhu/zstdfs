import os
import sys
import requests
import json
import random
import time
import mimetypes
import hashlib
from requests_toolbelt import MultipartEncoder
import pickle

user_apikey={}

url_list_tags = "http://192.168.0.108:9090/api/list/tags"
url_list_caption = "http://192.168.0.108:9090/api/list/caption"
user_apikey['harry'] = '3835573875204703656'

#url_list_tags = "http://127.0.0.1:9090/api/list/tags"
#url_list_caption = "http://127.0.0.1:9090/api/list/caption"
#user_apikey['harry'] = '14125811004486689209'

user_keys = user_apikey.keys()

sess = requests.session()
sess.keep_alive = False
sess.mount('https://', requests.adapters.HTTPAdapter(pool_connections=100, pool_maxsize=200))
sess.mount('http://', requests.adapters.HTTPAdapter(pool_connections=100, pool_maxsize=200))

def list_tags(username):
	req = {}
	req["fuser"] = username
	req["fapikey"] = user_apikey[username]
	resp = sess.post(url_list_tags, data=req)
	if resp.status_code == 200:
		print(resp.text)
		return resp.json()
	else:
		print(resp.status_code)
		return {}

def list_caption(username,lang):
	req = {}
	req["fuser"] = username
	req["fapikey"] = user_apikey[username]
	req["flanguage"] = lang
	resp = sess.post(url_list_caption, data=req)
	if resp.status_code == 200:
		print(resp.text)
		return resp.json()
	else:
		print(resp.status_code)
		return {}

tagCountList = list_tags("harry")
#print(f'TAGS: {tagCountList}')
with open("zstdfs_my_tags.txt","w") as fw:
	for k,v in tagCountList.items():
		if v > 9:
			fw.write(k+"\n")



captionCountList = list_caption("harry","en")
# print(f'CAPTION_EN: {captionCountList}')
caps = []
with open("zstdfs_my_caption_en.txt","w") as fw:
	for k,v in captionCountList.items():
		if v > 9 and k.count(" ") < 2:
			caps.append(k)
	print(caps)
	#caps = list(set(caps))
	for cap in caps:
		fw.write(cap+": \n")


captionCountList = list_caption("harry","cn")
# print(f'CAPTION_CN: {captionCountList}')
with open("zstdfs_my_caption_cn.txt","w") as fw:
	for k,v in captionCountList.items():
		if v > 9:
			fw.write(k+"\n")


