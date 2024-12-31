#! /bin/bash
. .env
. common.sh
. check-env.sh


function pre_longhorn_iscsi() {
  tee $KUBESPRAY_CACHE/kubespray-$KUBESPRAY_VERSION/longhorn-iscsi-installation.yaml >> /dev/null <<EOF
---
- name: Install packages and load modules
  hosts: all
  become: yes

  tasks:
    - name: Install packages on Debian/Ubuntu
      ansible.builtin.package:
        name:
          - open-iscsi
        state: present
      when: ansible_distribution == 'Debian' or ansible_distribution == 'Ubuntu'

    - name: Install packages on CentOS/RHEL
      ansible.builtin.package:
        name:
          - iscsi-initiator-utils
        state: present
      when: ansible_distribution == 'CentOS' or ansible_distribution == 'RedHat'

    - name: Configure initiator name on Debian/Ubuntu
      ansible.builtin.shell: echo "InitiatorName=$(/sbin/iscsi-iname)" > /etc/iscsi/initiatorname.iscsi
      when: ansible_distribution == 'Debian' or ansible_distribution == 'Ubuntu'

    - name: Configure initiator name on CentOS/RHEL
      ansible.builtin.shell: echo "InitiatorName=$(/sbin/iscsi-iname)" > /etc/iscsi/initiatorname.iscsi
      when: ansible_distribution == 'CentOS' or ansible_distribution == 'RedHat'

    - name: Enable and start iscsid on Debian/Ubuntu
      ansible.builtin.systemd:
        name: iscsid
        state: started
        enabled: yes
      when: ansible_distribution == 'Debian' or ansible_distribution == 'Ubuntu'

    - name: Enable and start iscsid on CentOS/RHEL
      ansible.builtin.systemd:
        name: iscsid
        state: started
        enabled: yes
      when: ansible_distribution == 'CentOS' or ansible_distribution == 'RedHat'

    - name: Load module on Debian/Ubuntu
      ansible.builtin.shell: modprobe iscsi_tcp
      when: ansible_distribution == 'Debian' or ansible_distribution == 'Ubuntu'

    - name: Load module on CentOS/RHEL
      ansible.builtin.shell: modprobe iscsi_tcp
      when: ansible_distribution == 'CentOS' or ansible_distribution == 'RedHat'

    - name: Install packages on Kylin V10
      ansible.builtin.package:
        name:
          - iscsi-initiator-utils
        state: present
      when: ansible_distribution == 'Kylin' and ansible_distribution_version == 'v10'

    - name: Configure iSCSI initiator name on Kylin V10
      ansible.builtin.shell: echo "InitiatorName=$(/sbin/iscsi-iname)" > /etc/iscsi/initiatorname.iscsi
      when: ansible_distribution == 'Kylin' and ansible_distribution_version == 'v10'

    - name: Enable and start iscsid service on Kylin V10
      ansible.builtin.shell: systemctl enable iscsid && systemctl start iscsid
      when: ansible_distribution == 'Kylin' and ansible_distribution_version == 'v10'

    - name: Load module on Kylin V10
      ansible.builtin.shell: modprobe iscsi_tcp
      when: ansible_distribution == 'Kylin' and ansible_distribution_version == 'v10'

    - name: Install packages on UOS 20
      ansible.builtin.package:
        name:
          - open-iscsi
        state: present
      when: ansible_distribution == 'UOS' and ansible_distribution_version == '20'

    - name: Configure iSCSI initiator name on UOS 20
      ansible.builtin.shell: echo "InitiatorName=$(/sbin/iscsi-iname)" > /etc/iscsi/initiatorname.iscsi
      when: ansible_distribution == 'UOS' and ansible_distribution_version == '20'

    - name: Enable and start iscsid service on UOS 20
      ansible.builtin.shell: systemctl enable iscsid && systemctl start iscsid
      when: ansible_distribution == 'UOS' and ansible_distribution_version == '20'

    - name: Load module on UOS 20
      ansible.builtin.shell: modprobe iscsi_tcp
      when: ansible_distribution == 'UOS' and ansible_distribution_version == '20'

    - name: Install packages on SUSE
      ansible.builtin.package:
        name:
          - open-iscsi
        state: present
      when: ansible_distribution == 'SUSE'

    - name: Configure iSCSI initiator name on SUSE
      ansible.builtin.shell: echo "InitiatorName=$(/sbin/iscsi-iname)" > /etc/iscsi/initiatorname.iscsi
      when: ansible_distribution == 'SUSE'

    - name: Enable and start iscsid service on SUSE
      ansible.builtin.shell: systemctl enable iscsid && systemctl start iscsid
      when: ansible_distribution == 'SUSE'

    - name: Load module on SUSE
      ansible.builtin.shell: modprobe iscsi_tcp
      when: ansible_distribution == 'SUSE'
EOF
  check_docker
  docker_ps=$(docker ps | grep "$KUBESPRAY_NAME")
  if [ -n "$docker_ps" ]; then
      echo "Error: Docker $KUBESPRAY_NAME already exists"
      exit 0
  fi
  kubespray_image=$KUBESPRAY_NAME:v$KUBESPRAY_VERSION
  tar_name=$(echo ${kubespray_image##*/} | sed s/:/-/g).tar
  echo "===>loading kubespray_image"
  docker load -i $IMAGES_OUTPUT/$tar_name
  echo "===> starting kubespray"
  docker run --rm -it -v $KUBESPRAY_CACHE/kubespray-$KUBESPRAY_VERSION/:/kubespray/ -v /root/.ssh/id_rsa:/root/.ssh/id_rsa --name $KUBESPRAY_NAME $kubespray_image \
      ansible-playbook -i /kubespray/inventory/$CLUSTER_NAME/hosts.yaml --private-key /root/.ssh/id_rsa --become --become-user=root longhorn-iscsi-installation.yaml -vvvvv
  if [ $? -ne 0 ]; then
    echo "Error: Failed to run Docker kubespray container."
    exit 1
  else
    echo "Create kubespray done."
  fi
}

function pre_longhorn_nfs() {
  tee $KUBESPRAY_CACHE/kubespray-$KUBESPRAY_VERSION/longhorn-nfs-installation.yaml >> /dev/null <<EOF
---
- name: Install packages and load modules
  hosts: all
  become: yes

  tasks:
    - name: Install packages on Debian/Ubuntu
      ansible.builtin.package:
        name:
          - nfs-common
        state: present
      when: ansible_distribution == 'Debian' or ansible_distribution == 'Ubuntu'

    - name: Install packages on CentOS/RHEL
      ansible.builtin.package:
        name:
          - nfs-utils
        state: present
      when: ansible_distribution == 'CentOS' or ansible_distribution == 'RedHat'

    - name: Install packages on Kylin V10
      ansible.builtin.package:
        name:
          - nfs-utils
        state: present
      when: ansible_distribution == 'Kylin' and ansible_distribution_version == 'v10'

    - name: Install packages on UOS20
      ansible.builtin.package:
        name:
          - nfs-common
        state: present
      when: ansible_distribution == 'UOS' and ansible_distribution_version == '20'

    - name: Install packages on SUSE
      ansible.builtin.package:
        name:
          - nfs-client
        state: present
      when: ansible_distribution == 'SUSE'
EOF
  check_docker
  docker_ps=$(docker ps | grep "$KUBESPRAY_NAME")
  if [ -n "$docker_ps" ]; then
      echo "Error: Docker $KUBESPRAY_NAME already exists"
      exit 0
  fi
  kubespray_image=$KUBESPRAY_NAME:v$KUBESPRAY_VERSION
  tar_name=$(echo ${kubespray_image##*/} | sed s/:/-/g).tar
  echo "===>loading kubespray_image"
  docker load -i $IMAGES_OUTPUT/$tar_name
  echo "===> starting kubespray"
  docker run --rm -it -v $KUBESPRAY_CACHE/kubespray-$KUBESPRAY_VERSION/:/kubespray/ -v /root/.ssh/id_rsa:/root/.ssh/id_rsa --name $KUBESPRAY_NAME $kubespray_image \
      ansible-playbook -i /kubespray/inventory/$CLUSTER_NAME/hosts.yaml --private-key /root/.ssh/id_rsa --become --become-user=root longhorn-nfs-installation.yaml -vvvvv
  if [ $? -ne 0 ]; then
    echo "Error: Failed to run Docker kubespray container."
    exit 1
  else
    echo "Create kubespray done."
  fi
}

function pre_hugepage() {
  tee $KUBESPRAY_CACHE/kubespray-$KUBESPRAY_VERSION/longhorn-hugepages-installation.yaml >> /dev/null <<EOF
---
- name: Install packages and load modules
  hosts: all
  become: yes

  tasks:
    - name: Install packages on Debian/Ubuntu
      ansible.builtin.package:
        name:
          - linux-modules-extra-`uname -r`
          - nvme-cli
        state: present
      when: ansible_distribution == 'Debian' or ansible_distribution == 'Ubuntu'

    - name: Install packages on CentOS/RHEL
      ansible.builtin.package:
        name:
          - nvme-cli
        state: present
      when: ansible_distribution == 'CentOS' or ansible_distribution == 'RedHat'

    - name: Install packages on Kylin V10
      ansible.builtin.package:
        name:
          - nvme-cli
        state: present
      when: ansible_distribution == 'Kylin' and ansible_distribution_version == 'v10'

    - name: Install packages on UOS20
      ansible.builtin.package:
        name:
          - linux-modules-extra-`uname -r`
          - nvme-cli
        state: present
      when: ansible_distribution == 'UOS' and ansible_distribution_version == '20'

    - name: Install packages on SUSE
      ansible.builtin.package:
        name:
          - nvme-cli
        state: present
      when: ansible_distribution == 'SUSE'

    - name: Load module on Debian/Ubuntu
      ansible.builtin.shell: modprobe uio && modprobe uio_pci_generic && modprobe nvme-tcp
      when: ansible_distribution == 'Debian' or ansible_distribution == 'Ubuntu'

    - name: Load module on CentOS/RHEL
      ansible.builtin.shell: modprobe uio && modprobe uio_pci_generic && modprobe nvme-tcp
      when: ansible_distribution == 'CentOS' or ansible_distribution == 'RedHat'

    - name: Load module on KylinV10
      ansible.builtin.shell: modprobe uio && modprobe uio_pci_generic && modprobe nvme-tcp
      when: ansible_distribution == 'Kylin' and ansible_distribution_version == 'v10'

    - name: Load module on UOS
      ansible.builtin.shell: modprobe uio && modprobe uio_pci_generic && modprobe nvme-tcp
      when: ansible_distribution == 'UOS' and ansible_distribution_version == '20'

    - name: Load module on SUSE
      ansible.builtin.shell: modprobe uio && modprobe uio_pci_generic && modprobe nvme-tcp
      when: ansible_distribution == 'SUSE'

    - name: Configure hugepages
      ansible.builtin.shell: echo 1024 > /sys/kernel/mm/hugepages/hugepages-2048kB/nr_hugepages

    - name: Configure hugepages permanent
      ansible.builtin.shell: echo "vm.nr_hugepages=1024" >> /etc/sysctl.conf

    - name: Make hugepages take effect
      ansible.builtin.shell: sysctl -p /etc/sysctl.conf
EOF
  check_docker
  docker_ps=$(docker ps | grep "$KUBESPRAY_NAME")
  if [ -n "$docker_ps" ]; then
      echo "Error: Docker $KUBESPRAY_NAME already exists"
      exit 0
  fi
  kubespray_image=$KUBESPRAY_NAME:v$KUBESPRAY_VERSION
  tar_name=$(echo ${kubespray_image##*/} | sed s/:/-/g).tar
  echo "===>loading kubespray_image"
  docker load -i $IMAGES_OUTPUT/$tar_name
  echo "===> starting kubespray"
  docker run --rm -it -v $KUBESPRAY_CACHE/kubespray-$KUBESPRAY_VERSION/:/kubespray/ -v /root/.ssh/id_rsa:/root/.ssh/id_rsa --name $KUBESPRAY_NAME $kubespray_image \
      ansible-playbook -i /kubespray/inventory/$CLUSTER_NAME/hosts.yml --private-key /root/.ssh/id_rsa --become --become-user=root longhorn-hugepages-installation.yaml -vvvvv
  if [ $? -ne 0 ]; then
    echo "Error: Failed to run Docker kubespray container."
    exit 1
  else
    echo "Create kubespray done."
  fi
}