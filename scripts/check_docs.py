import pathlib,re
errors=[]
for p in pathlib.Path('.').rglob('*.md'):
 if any(v in p.parts for v in ['.git','dist','results']):continue
 for target in re.findall(r'\]\(([^)]+)\)',p.read_text()):
  if '://' in target or target.startswith('#'):continue
  target=target.split('#')[0]
  if target and not (p.parent/target).exists():errors.append(f'{p}: missing {target}')
for f in ['assets/demo.gif','assets/demo.svg','assets/banner.svg','assets/logo.svg','assets/architecture.svg']:
 if not pathlib.Path(f).exists():errors.append('missing '+f)
if errors:raise SystemExit('\n'.join(errors))
print('Local Markdown links and required assets verified.')
