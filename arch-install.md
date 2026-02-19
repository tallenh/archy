- Prepare
	-
	  ```
	  setfont ter-v32b
	  timedatectl set-ntp true
	  ```
- Partitioning
	-
	  ```
	  cfdisk
	  ```
		- gpt
		- new part, 512M, type=EFI System
		- new part, remaining, type=Linux filesystem
		- write, quit
	-
	  ```
	  mkfs.fat -F32 /dev/sda1
	  mkfs.btrfs -L ArchRoot /dev/sda2
	  ```
- Configuring Btrfs
	-
	  ```
	  mount /dev/sda2 /mnt
	  btrfs subvolume create /mnt/@
	  btrfs subvolume create /mnt/@home
	  btrfs subvolume create /mnt/@snapshots
	  btrfs subvolume create /mnt/@var_log
	  umount /mnt
	  mount -o noatime,compress=zstd,subvol=@ /dev/sda2 /mnt
	  mkdir -p /mnt/{boot,home,snapshots,var/log}
	  mount -o noatime,compress=zstd,subvol=@home /dev/sda2 /mnt/home
	  mount -o noatime,compress=zstd,subvol=@snapshots /dev/sda2 /mnt/snapshots
	  mount -o noatime,compress=zstd,subvol=@var_log /dev/sda2 /mnt/var/log
	  mount /dev/sda1 /mnt/boot
	  mkdir -p /mnt/etc
	  genfstab -U /mnt >> /mnt/etc/fstab
	  ```
- Base System Installation
	-
	  ```
	  pacstrap /mnt base linux linux-firmware sudo vim
	  arch-chroot /mnt
	  ```
- System Configuration (chroot)
	-
	  ```
	  timedatectl set-timezone America/Chicago
	  hwclock --systohc
	  locale-gen: Generate locales after editing the /etc/locale.gen file
	  sed -i '/en_US.UTF-8/s/^#//' /etc/locale.gen
	  echo "LANG=en_US.UTF-8" > /etc/locale.conf
	  echo "your-hostname" > /etc/hostname
	  passwd
	  useradd -m allenh
	  passwd allenh
	  usermod -aG wheel,audio,video,optical,storage,input allenh
	  ```
	- visudo: edit the privileges file to allow the use of sudo to the wheel group
	  uncomment one of the two options to allow wheel, either requiring a password or not  
- Swap, Bootloader, Servicesq
	-
	  ```
	  pacman -S zram-generator
	  vim /etc/systemd/zram-generator.conf
	  ```
		-
		  ```
		  [zram0]
		  zram-size = ram / 2
		  compression-algorithm = zstd
		  swap-prioirty = 100
		  fs-type = swap
		  ```
	-
	  ```
	  pacman -S grub efibootmgr
	  mkdir -p /boot/efi
	  mount /dev/sda1 /boot/efi
	  grub-install --target=x86_64-efi --efi-directory=/boot/efi --bootloader-id=GRUB
	  grub-mkconfig -o /boot/grub/grub.cfg
	  pacman -S networkmanager
	  ```
- Software and AUR Installation
	-
	  ```
	  pacman -S base-devel git
	  su allenh
	  git clone https://aur.archlinux.org/yay.git
	  cd yay
	  makepkg -si
	  exit/exit
	  umount -l /mnt
	  reboot
	  ```
- Post Reboot
	-
	  ```
	  sudo systemctl start /dev/zram0
	  sudo systemctl enable --now NetworkManager
	  ```
- Gnome
	-
	  ```
	  pacman -S gnome
	  systemctl enable gdm.service
	  reboot
	  ```
	-
