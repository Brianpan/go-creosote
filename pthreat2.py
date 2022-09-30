import tarfile
import os

filename = "test123"
temp_dir = "~/"

if filename.endwith("tar"):
  obj = tarfile.open(filename, "r")
  extracted_file_names = []
  for j,member in enumerate(obj.getmembers()):
    obj.extract(member, temp_dir)
    extracted_file_name = os.path.join(temp_dir, obj.getnames()[j])
    extracted_file_names.append(extracted_file_name)
