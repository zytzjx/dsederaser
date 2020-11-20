import os
import redis
import syslog
import zipfile
import stat
import json

syslog.openlog('dsed.deployment')
syslog.syslog('dsed.deployment: ++ start')

r = redis.Redis()

dsed_home = os.getenv("DSEDHOME", '')
if not bool(dsed_home):
    dsed_home = '/opt/futuredial/dsed'
    os.putenv("DSEDHOME", dsed_home)

hydradownload_running = r.get('hydradownload.running')
hydradownload_status = r.get('hydradownload.status')
hydradownload_clientstatus = r.get('hydradownload.clientstatus')

syslog.syslog('dsed.deployment: hydradownload.running={}'.format(hydradownload_running))
syslog.syslog('dsed.deployment: hydradownload.status={}'.format(hydradownload_status))
syslog.syslog('dsed.deployment: hydradownload.clientstatus={}'.format(hydradownload_clientstatus))

def stopService():
    #stop service
    r.publish("inittask", json.dumps({'message':'stop daemon service...'}))
    sudoPassword = 'qa'
    command = 'systemctl stop dseddetect.service'
    os.system('echo %s|sudo -S %s' % (sudoPassword, command))
    command = 'systemctl stop hdderaser.service'
    os.system('echo %s|sudo -S %s' % (sudoPassword, command))

def startService():
    #stop service
    r.publish("inittask", json.dumps({'message':'start daemon service...'}))
    sudoPassword = 'qa'
    command = 'systemctl start dseddetect.service'
    os.system('echo %s|sudo -S %s' % (sudoPassword, command))
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


if hydradownload_running==b'0' and hydradownload_status==b'complete':
    syslog.syslog('dsed.deployment: start deployment ...')
    syslog.syslog('dsed.deployment: set hydradownload.status=pause')
    r.set('hydradownload.status', 'pause')
    # keys = ['hydradownload.framework', 'hydradownload.phonedll']
    # hydradownload.framework
    framework_ok = True
    syslog.syslog('dsed.deployment: read key hydradownload.framework')
    i = r.spop('hydradownload.framework')
    stopService()
    while bool(i):
        fn = i.decode('utf-8')
        syslog.syslog('dsed.deployment: value {}'.format(fn))
        try:
            if os.path.exists(fn):
                with zipfile.ZipFile(fn, 'r') as f:
                    f.extractall(os.environ['DSEDHOME'])
                os.remove(fn)
        except:
            syslog.syslog('dsed.deployment: exception')
            framework_ok = False
        i = r.spop('hydradownload.framework')
        pass

    changefileExe()
    startService()
    
    # hydradownload.phonedll
    syslog.syslog('dsed.deployment: read key hydradownload.phonedll')
    # hydradownload.
    syslog.syslog('dsed.deployment: read key hydradownload.phonetips')
    # save hydradownload.clientstatus
    if framework_ok :
        fn = os.path.join(os.environ['DSEDHOME'], 'clientstatus.json')
        with open(fn, 'w') as f:
            f.write(hydradownload_clientstatus.decode('utf-8'))

syslog.syslog('dsed.deployment: delete hydradownload keys')
for k in r.scan_iter('hydradownload*'):
    r.delete(k)
syslog.syslog('dsed.deployment: -- complete')
syslog.closelog()
r.publish("inittask", '{"action":"close"}')