import os
import json
import redis
import syslog
import subprocess

import time

syslog.openlog('dsed.autoupdater')
r = redis.Redis()
# downloader
dsed_home = os.getenv("DSEDHOME", '/opt/futuredial/dsed')
os.putenv('DSEDHOME', dsed_home)
syslog.syslog('autoupdater: start downlaoding... {}'.format(dsed_home))
fn = os.path.join(dsed_home, 'hydradownload')
syslog.syslog('autoupdater: start downloand... {} '.format(fn))
r.publish("inittask", json.dumps({'message':'start download...'}))
if os.path.exists(fn):
    p = subprocess.Popen([fn], cwd=dsed_home)
    p.wait()
    syslog.syslog('autoupdater: hydradownload return: {}'.format(p.returncode))
syslog.syslog('autoupdater: complete downloand.')

# deploy
syslog.syslog('autoupdater: start deployment ...')
r.publish("inittask", json.dumps({'message':'start deployment...'}))
dsed_status = r.get('dsed.status')
if bool(dsed_status) and dsed_status.decode('utf-8') == 'running':
    syslog.syslog('autoupdater: deployment postponed ...')
else:
    syslog.syslog('autoupdater: start deployment ...')
    time.sleep(1)
    fn = os.path.join(dsed_home, 'cmcdeployment.py')
    if os.path.exists(fn):
        p = subprocess.Popen(['python3', fn], cwd=dsed_home)
        p.wait()
syslog.syslog('autoupdater: deployment complete ...')
syslog.closelog()