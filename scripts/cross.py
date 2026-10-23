"""Cross-compile; optionally archive and checksum binaries without publishing."""
import os, subprocess, pathlib, sys, tarfile, zipfile, hashlib
out=pathlib.Path('dist'); out.mkdir(exist_ok=True)
for system in ['linux','darwin','windows']:
 for arch in ['amd64','arm64']:
  name=f'reconcilelab-{system}-{arch}'+('.exe' if system=='windows' else '')
  target=out/name
  subprocess.run(['go','build','-trimpath','-o',str(target),'./cmd/reconcilelab'],env={**os.environ,'GOOS':system,'GOARCH':arch,'CGO_ENABLED':'0'},check=True)
  if '--archive' in sys.argv:
   if system=='windows':
    with zipfile.ZipFile(str(target)+'.zip','w',zipfile.ZIP_DEFLATED) as f: f.write(target,'reconcilelab.exe');f.write('LICENSE','LICENSE')
   else:
    with tarfile.open(str(target)+'.tar.gz','w:gz') as f: f.add(target,arcname='reconcilelab');f.add('LICENSE',arcname='LICENSE')
files=sorted(p for p in out.iterdir() if p.name!='SHA256SUMS')
(out/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.name+'\n' for p in files))
print(f'Compiled six targets; checksums in {out}/SHA256SUMS. Native execution not implied.')
