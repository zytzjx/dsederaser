#!/bin/bash -xv
set -e

# echo "This script is running as root $SUDO_USER"

if [[ $(lsb_release -rs) != "20.04" ]]; then
   echo "Non-compatible version"
   exit 2
fi

sudo mkdir -p /opt/futuredial
sudo chown $USER:$USER /opt/futuredial
sudo mkdir -p /opt/futuredial/dsed
sudo mkdir -p /opt/futuredial/hydradownloader
sudo chown $USER:$USER /opt/futuredial/dsed
sudo chown $USER:$USER /opt/futuredial/hydradownloader

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
serialnumber=""
while read -r -n 1 key; do
   if [[ $key == "" ]]; then
      break
   fi
   # Add the key to the variable which is pressed by the user.
   serialnumber+=$key
done
echo $serialnumber

echo "start downloading dsedcmc"
wget https://bitbucket.org/yuetingzhang/dsedcmc/raw/799ecc5033659233441261f33df6caa842e1d26b/dsedcmc -O dsedcmc

cp ./dsedcmc $DSEDHOME/dsedcmc
chmod +x $DSEDHOME/dsedcmc

sudo apt install ssh redis -y
sudo apt install smartmontools -y
sudo apt install wxhexeditor -y
sudo apt install lsscsi -y
#sudo apt install openssh-server -y
#gsettings set org.gnome.Vino require-encryption false

sudo apt install python3-pip -y
pip3 install redis
pip3 install pyqt5 
pip3 install pyqt5 --upgrade

sudo sed -i 's/databases 16/databases 81/g' /etc/redis/redis.conf
sudo systemctl restart redis.service

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

echo "start downloading hydradownload"
wget https://github.com/zytzjx/hydradownload/raw/master/hydradownload -O hydradownload
chmod +x hydradownload

wget -i request.txt

# url, servicename
InstallService(){
   sname=/etc/systemd/system/$2
   wget $1 -O aaa.service
   sudo mv ./aaa.service $sname
   sudo chmod 644 $sname
   sudo systemctl daemon-reload
   sudo systemctl enable $2
   sudo systemctl start $2
}

InstallService https://raw.githubusercontent.com/zytzjx/dsederaser/master/hdderaser.service hdderaser.service
#InstallService https://raw.githubusercontent.com/zytzjx/dsederaser/master/hdderaser.service hdderaser.service

InstallShortcut(){
   cp $DSEDHOME/dsed.desktop ~/Desktop/dsed.desktop
   chmod +0744 ~/Desktop/dsed.desktop
   gio set ~/Desktop/dsed.desktop "metadata::trusted" true
}