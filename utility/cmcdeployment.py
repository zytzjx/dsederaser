import os
import sys
import redis
import logging
import zipfile
import stat
import time
from logging.handlers import RotatingFileHandler



dsed_home = os.getenv("DSEDHOME", '')
if not bool(dsed_home):
    dsed_home = '/opt/futuredial/dsed'
    os.putenv("DSEDHOME", dsed_home)

log_formatter = logging.Formatter('%(asctime)s %(levelname)s %(name)s(%(lineno)d) %(message)s')
logFile = os.path.join(dsed_home, 'cmcdeployment.log')
my_handler = RotatingFileHandler(logFile, mode='a', maxBytes=50*1024*1024, backupCount=2, encoding=None, delay=0)
my_handler.setFormatter(log_formatter)                                 
my_handler.setLevel(logging.INFO)
log = logging.getLogger('cmcdeployment')
log.setLevel(logging.INFO)
log.addHandler(my_handler)
log.addHandler(logging.StreamHandler(sys.stdout))
log.info('dsed.deployment')
log.info('dsed.deployment: ++ start')

r = redis.Redis()


hydradownload_running = r.get('hydradownload.running')
hydradownload_status = r.get('hydradownload.status')
hydradownload_clientstatus = r.get('hydradownload.clientstatus')

log.info('dsed.deployment: hydradownload.running={}'.format(hydradownload_running))
log.info('dsed.deployment: hydradownload.status={}'.format(hydradownload_status))
log.info('dsed.deployment: hydradownload.clientstatus={}'.format(hydradownload_clientstatus))

def stopService():
    #stop service
    r.publish("inittask", '{"message":"stop service ..."}')
    sudoPassword = 'qa'
    command = 'systemctl stop dseddetect.service'
    os.system('echo %s|sudo -S %s' % (sudoPassword, command))
    command = 'systemctl stop hdderaser.service'
    os.system('echo %s|sudo -S %s' % (sudoPassword, command))

def startService():
    #stop service
    r.publish("inittask", '{"message":"start service ..."}')
    print("start dseddetect.service")
    sudoPassword = 'qa'
    command = 'systemctl start dseddetect.service'
    os.system('echo %s|sudo -S %s' % (sudoPassword, command))
    time.sleep(1)
    print("start hdderaser.service")
    command = 'systemctl start hdderaser.service'
    os.system('echo %s|sudo -S %s' % (sudoPassword, command))

# change file executable
def changefileExe():
    filenames=[
        "dsedcmc",
        "athenasetting",
        "dseddetect",
        "dsederaser",
        "dsedtransaction",
        "dskwipe",
        "sas2ircu"
    ]

    for filename in filenames:
        # chmod +x transaction
        fn = os.path.join(dsed_home,filename)
        if os.path.exists(fn):
            st = os.stat(fn)
            os.chmod(fn, st.st_mode | stat.S_IEXEC|stat.S_IXUSR|stat.S_IXGRP|stat.S_IXOTH)


framework_ok = False
if hydradownload_running==b'0' and hydradownload_status==b'complete':
    log.info('dsed.deployment: start deployment ...')
    log.info('dsed.deployment: set hydradownload.status=pause')
    r.set('hydradownload.status', 'pause')
    # keys = ['hydradownload.framework', 'hydradownload.phonedll']
    # hydradownload.framework
    log.info('dsed.deployment: read key hydradownload.framework')
    i = r.get('hydradownload.framework')
    if bool(i):
        stopService()
        time.sleep(2)
        
    if  bool(i):
        fn = i.decode('utf-8')
        log.info('dsed.deployment: value {}'.format(fn))
        try:
            if os.path.exists(fn):
                with zipfile.ZipFile(fn, 'r') as f:
                    f.extractall(os.environ['DSEDHOME'])
                os.remove(fn)

                framework_ok = True
        except:
            log.info('dsed.deployment: exception')
            framework_ok = False
        #i = r.spop('hydradownload.framework')
        r.set('hydradownload.framework','')
        pass

    changefileExe()
    
    # hydradownload.phonedll
    log.info('dsed.deployment: read key hydradownload.phonedll')
    # hydradownload.
    log.info('dsed.deployment: read key hydradownload.phonetips')
    # save hydradownload.clientstatus
    if framework_ok :
        fn = os.path.join(os.environ['DSEDHOME'], 'clientstatus.json')
        with open(fn, 'w') as f:
            f.write(hydradownload_clientstatus.decode('utf-8'))

    
startService()

if framework_ok:
    log.info('dsed.deployment: delete hydradownload keys')
    for k in r.scan_iter('hydradownload*'):
        r.delete(k)

log.info('dsed.deployment: -- complete')
r.publish("inittask", '{"action":"close"}')