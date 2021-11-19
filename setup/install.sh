#!/bin/bash -xv
set -e

# echo "This script is running as root $SUDO_USER"

if [[ $(lsb_release -rs) != "20.04" ] && [ $(lsb_release -rs) != "18.04" ]]; then
   echo "Non-compatible version"
   exit 2
fi

echo -e "\e[1;33mPlease input [sudo] password:\e[0m"
{
   password=""
   while read -r -n 1 key; do
      if [[ $key == "" ]]; then
         break
      fi
      # Add the key to the variable which is pressed by the user.
      password+=$key
   done
} 2>/dev/null


echo $password | sudo -S mkdir -p /opt/futuredial
echo $password | sudo -S chown $USER:$USER /opt/futuredial
echo $password | sudo -S mkdir -p /opt/futuredial/dsed
echo $password | sudo -S mkdir -p /opt/futuredial/hydradownloader
echo $password | sudo -S chown $USER:$USER /opt/futuredial/dsed
echo $password | sudo -S chown $USER:$USER /opt/futuredial/hydradownloader

echo $password > /opt/futuredial/dsed/spusdwo

# echo add environment
if [[ -z $DSEDHOME ]]; then 
   echo "set DSEDHOME=/opt/futuredial/dsed"
   export DSEDHOME=/opt/futuredial/dsed
   echo 'export DSEDHOME=/opt/futuredial/dsed' >> ~/.bashrc
   source ~/.bashrc
fi

echo $DSEDHOME

# echo "input serial number" db6a1e7f-69f7-4e86-9d50-5c4b0abcbf08
# echo -e "\e[1;31mThis is red text\e[0m"
# echo "Please input this product SN:"
echo -e "\e[1;31mPlease input this product SN:\e[0m"
{
   serialnumber=""
   while read -r -n 1 key; do
      if [[ $key == "" ]]; then
         break
      fi
      # Add the key to the variable which is pressed by the user.
      serialnumber+=$key
   done
} 2>/dev/null
echo -e "\e[1;34mSERIAL:"$serialnumber"\e[0m"


echo $password | sudo -S apt update
#################################################################################
##################### ubnutu 18.04 install python3.8 ############################
## $ sudo apt-get update
## $ sudo apt-get install software-properties-common
## $ sudo add-apt-repository ppa:deadsnakes/ppa
## $ sudo apt-get update
## $ sudo apt-get install python3.8
## $ sudo rm /usr/bin/python3   
## $ sudo ln -s /usr/bin/python3.8 /usr/bin/python3
#################
## $ sudo apt install python3.8 
#################################################################################
echo "start downloading dsedcmc"
wget https://github.com/zytzjx/dsederaser/raw/master/setup/dsedcmc -O dsedcmc

cp ./dsedcmc $DSEDHOME/dsedcmc
chmod +x $DSEDHOME/dsedcmc

#sudo apt install ssh redis -y
echo $password | sudo -S apt install redis -y
echo $password | sudo -S apt install smartmontools -y
#sudo apt install wxhexeditor -y
echo $password | sudo -S apt install lsscsi -y
echo $password | sudo -S apt install python3-pyqt5 -y

#sudo apt install openssh-server -y
#gsettings set org.gnome.Vino require-encryption false

echo $password | sudo -S apt install python3-pip -y
python3 -m pip install --upgrade pip
pip3 install --upgrade pip
python3 -m pip install --upgrade setuptools
pip3 install PyYAML==6.0 
pip3 install redis
pip3 install pyqt5==5.15.4 
pip3 install pyqt5 --upgrade

#remove office
echo $password | sudo -S apt-get remove --purge libreoffice* -y
echo $password | sudo -S apt-get clean -y
echo $password | sudo -S apt-get autoremove -y

echo $password | sudo -S sed -i 's/databases 16/databases 81/g' /etc/redis/redis.conf
echo $password | sudo -S systemctl restart redis.service

#download 
echo "start downloading the CMC config..."
cd $DSEDHOME
$DSEDHOME/dsedcmc -uuid=$serialnumber
if [ $? -eq 0 ]; then
  echo "Success: Serial Number is verified."
else
  echo "Failure: Serial Number can not be verify." >&2
  exit 3
fi

# disable autoupdate
cat > 20auto-upgrades << "EOF2"
APT::Periodic::Update-Package-Lists "0";
APT::Periodic::Download-Upgradeable-Packages "0";
APT::Periodic::AutocleanInterval "0";
APT::Periodic::Unattended-Upgrade "1";
EOF2

echo $password | sudo -S cp 20auto-upgrades /etc/apt/apt.conf.d/20auto-upgrades


echo "start downloading hydradownload"
wget https://github.com/zytzjx/hydradownload/raw/master/hydradownload -O hydradownload
chmod +x hydradownload

wget https://raw.githubusercontent.com/zytzjx/dsederaser/master/utility/autoupdater.py -O autoupdater.py
wget https://raw.githubusercontent.com/zytzjx/dsederaser/master/utility/cmcdeployment.py -O cmcdeployment.py

python3 autoupdater.py
python3 cmcdeployment.py
crontab $DSEDHOME/download_cron
#wget -i request.txt

InstallShortcut(){
   cp $DSEDHOME/dsed.desktop ~/Desktop/dsed.desktop
   chmod +x ~/Desktop/dsed.desktop
   gio set ~/Desktop/dsed.desktop "metadata::trusted" true
}
InstallShortcut


# url, servicename
InstallService(){
   sname=/etc/systemd/system/$1
   #wget $2 -O aaa.service
   echo $password | sudo -S mv ./$1 $sname
   echo $password | sudo -S chmod 644 $sname
   echo $password | sudo -S systemctl daemon-reload
   if echo $password | sudo -S systemctl enable $1; then
       echo "enable failed, reboot system"
   fi
   # if echo $password | sudo -S systemctl start $1; then
   #     echo "start failed"
   # fi
}
#https://raw.githubusercontent.com/zytzjx/dsederaser/master/hdderaser.service
InstallService hdderaser.service 
#https://raw.githubusercontent.com/zytzjx/dseddetect/master/dseddetect.service 
InstallService dseddetect.service 

echo -e "\e[1;34mReboot SYSTEM\e[0m"
