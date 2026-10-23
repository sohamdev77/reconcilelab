"""Record real CLI subprocess output, then render a low-frame-rate terminal GIF.
Requires Pillow. No invented terminal output. Commands and exit codes are saved.
"""
from pathlib import Path
import subprocess, html, json
from PIL import Image, ImageDraw, ImageFont
commands=[(['version'],0),(['test','examples/basic-controller','--output','results/demo-buggy'],1),(['replay','results/demo-buggy/counterexample.yaml','--output','results/demo-replay'],1),(['replay','results/demo-buggy/counterexample.yaml','--adapter','fixed','--output','results/demo-fixed'],0)]
font=ImageFont.truetype('/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf',17)
frames=[];record=[];pages=[]
for args,code in commands:
 run=subprocess.run(['./bin/reconcilelab',*args],text=True,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
 if run.returncode!=code:raise SystemExit(f'Unexpected code {run.returncode}: {run.stdout}')
 command='$ reconcilelab '+' '.join(args)
 # Wrap command text for display; raw transcript retains the original.
 import textwrap
 lines=textwrap.wrap(command,100,subsequent_indent='  ')+['']+run.stdout.rstrip().splitlines()+['',f'[exit {run.returncode}]']
 record.append(command+'\n'+run.stdout+f'[exit {run.returncode}]\n')
 pages.append(lines)
 for shown in range(1,len(lines)+1):
  image=Image.new('RGB',(1100,450),'#0B1020');d=ImageDraw.Draw(image)
  d.rounded_rectangle((12,12,1088,438),radius=14,outline='#334155',width=2)
  d.text((30,26),'●  ●  ●     ReconcileLab / local simulation',font=font,fill='#A78BFA')
  for i,line in enumerate(lines[:shown]):
   color='#7DD3FC' if line.startswith('$') else '#FB7185' if 'VIOLATION' in line else '#34D399' if 'PASS' in line else '#E5E7EB'
   d.text((30,70+i*24),line,font=font,fill=color)
  frames.append(image)
 # Hold each complete page for 3 seconds at a low effective frame rate.
 for _ in range(6):frames.append(image.copy())
Path('assets/demo.txt').write_text('\n'.join(record))
frames[0].save('assets/demo.gif',save_all=True,append_images=frames[1:],duration=450,loop=0,optimize=True)
# Static fallback shows all command output, readable without motion.
all_lines=[]
for lines in pages:all_lines+=lines+['']
height=70+24*len(all_lines)
svg=f'<svg xmlns="http://www.w3.org/2000/svg" width="1100" height="{height}" viewBox="0 0 1100 {height}"><rect width="1100" height="{height}" rx="16" fill="#0B1020"/><g font-family="monospace" font-size="17">'
for i,line in enumerate(all_lines):
 color='#7DD3FC' if line.startswith('$') else '#FB7185' if 'VIOLATION' in line else '#34D399' if 'PASS' in line else '#E5E7EB'
 svg+=f'<text x="25" y="{35+i*24}" fill="{color}">{html.escape(line)}</text>'
Path('assets/demo.svg').write_text(svg+'</g></svg>')
print('Recorded four real CLI commands; GIF, SVG and raw transcript saved.')
