import os
import redis
import syslog
import zipfile

syslog.openlog('dsed.deployment')
syslog.syslog('dsed.deployment: ++ start')

r = redis.Redis()
hydradownload_running = r.get('hydradownload.running')
hydradownload_status = r.get('hydradownload.status')
hydradownload_clientstatus = r.get('hydradownload.clientstatus')

syslog.syslog('dsed.deployment: hydradownload.running={}'.format(hydradownload_running))
syslog.syslog('dsed.deployment: hydradownload.status={}'.format(hydradownload_status))
syslog.syslog('dsed.deployment: hydradownload.clientstatus={}'.format(hydradownload_clientstatus))

if hydradownload_running==b'0' and hydradownload_status==b'complete':
    syslog.syslog('dsed.deployment: start deployment ...')
    syslog.syslog('dsed.deployment: set hydradownload.status=pause')
    r.set('hydradownload.status', 'pause')
    # keys = ['hydradownload.framework', 'hydradownload.phonedll']
    # hydradownload.framework
    framework_ok = True
    syslog.syslog('dsed.deployment: read key hydradownload.framework')
    i = r.spop('hydradownload.framework')
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