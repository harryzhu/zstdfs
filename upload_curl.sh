auth=admin:123
filepath=/Users/harry/Downloads/d3.v7.js
fuser=admin
fgroup=bootstrap22
fprefix=prod/v3.4.5
server_host_port=localhost:8080
# upload
curl -X POST -H "Content-Type: multipart/form-data"  -F "file=@${filepath}" --user ${auth} -F "fuser=${fuser}" -F "fgroup=${fgroup}" -F "fprefix=${fprefix}" "http://${server_host_port}/admin/upload"