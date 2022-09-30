import tarfile

with tarfile.open("xxx", "r:gz") as tar:
  tar.extractall()
  print("extract")
