import os
import sys
import json
import redis
import logging
import subprocess
from logging.handlers import RotatingFileHandler
import time

dsed_home = os.getenv("DSEDHOME", '')
if not bool(dsed_home):
    dsed_home = '/opt/futuredial/dsed'
    os.putenv("DSEDHOME", dsed_home)

log_formatter = logging.Formatter(
    '%(asctime)s %(levelname)s %(name)s(%(lineno)d) %(message)s')

logFile = os.path.join(dsed_home, 'autoupdater.log')
my_handler = RotatingFileHandler(
    logFile, mode='a', maxBytes=50*1024*1024, backupCount=2, encoding=None, delay=0)
my_handler.setFormatter(log_formatter)
my_handler.setLevel(logging.INFO)
log = logging.getLogger('autoupdater')
log.setLevel(logging.INFO)
log.addHandler(my_handler)
log.addHandler(logging.StreamHandler(sys.stdout))

log.info('dsed.autoupdater')
r = redis.Redis()
# downloader
dsed_home = os.getenv("DSEDHOME", '/opt/futuredial/dsed')
os.putenv('DSEDHOME', dsed_home)
log.info('autoupdater: start downlaoding... {}'.format(dsed_home))
fn = os.path.join(dsed_home, 'hydradownload')
log.info('autoupdater: start downloand... {} '.format(fn))
r.publish("inittask", json.dumps({'message': 'start download...'}))
i = r.get('hydradownload.framework')
if not bool(i):
    if os.path.exists(fn):
        p = subprocess.Popen([fn], cwd=dsed_home)
        p.wait()
        log.info('autoupdater: hydradownload return: {}'.format(p.returncode))
log.info('autoupdater: complete downloand.')

# deploy
log.info('autoupdater: start deployment ...')
r.publish("inittask", json.dumps({'message': 'start deployment...'}))
dsed_status = r.get('dsed.status')
if bool(dsed_status) and dsed_status.decode('utf-8') == 'running':
    log.info('autoupdater: deployment postponed ...')
else:
    log.info('autoupdater: start deployment ...')
    time.sleep(1)
    fn = os.path.join(dsed_home, 'cmcdeployment.py')
    if os.path.exists(fn):
        p = subprocess.Popen(['python3', fn], cwd=dsed_home)
        p.wait()
log.info('autoupdater: deployment complete ...')
r.set("hydradownload.status", "idle")
r.close()
