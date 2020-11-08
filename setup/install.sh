#!/bin/bash -xv
set -e

# echo "This script is running as root $SUDO_USER"

if [[ $(lsb_release -rs) != "20.04" ]]; then
   echo "Non-compatible version"
   exit 2
fi

sudo mkdir -p /opt/futuredial
sudo chown $USER:$USER /opt/futuredial
sudo mkdir -p /opt/futuredial/hdses
sudo mkdir -p /opt/futuredial/hydradownloader
sudo chown $USER:$USER /opt/futuredial/hdses
sudo chown $USER:$USER /opt/futuredial/hydradownloader

# echo add environment
if [[ -z $HDSESHOME ]]; then 
   echo "set HDSESHOME=/opt/futuredial/hdses"
   export HDSESHOME=/opt/futuredial/hdses
   echo 'export HDSESHOME=/opt/futuredial/hdses' >> ~/.bashrc
   source ~/.bashrc
fi

echo $HDSESHOME

# echo "input serial number"
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

echo "start downloading anthenacmc"
wget https://github.com/zytzjx/anthenacmc/raw/master/anthenacmc -O anthenacmc

cp ./anthenacmc $HDSESHOME/anthenacmc
chmod +x $HDSESHOME/anthenacmc

sudo sed -i 's/databases 16/databases 81/g' /etc/redis/redis.conf
sudo apt install ssh redis -y
sudo apt install smartmontools -y
sudo apt install wxhexeditor -y
sudo apt install lsscsi -y
#sudo apt install openssh-server -y
#gsettings set org.gnome.Vino require-encryption false

sudo apt install python3-pip -y
sudo pip3 install redis
sudo pip3 install pyqt5 

#download 
echo "start downloading the CMC config..."
cd $HDSESHOME
$HDSESHOME/anthenacmc -uuid=$serialnumber
if [ $? -eq 0 ]; then
  echo "Success: Serial Number is verified."
else
  echo "Failure: Serial Number can not be verify." >&2
  exit 3
fi

echo "start downloading hydradownload"
wget https://github.com/zytzjx/hydradownload/raw/master/hydradownload -O hydradownload
chmod +x hydradownload




