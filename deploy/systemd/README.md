In order to use this service file you can either place it in `/etc/systemd/system` or in `~/.config/systemd/user/`.
I would recommend the later option and all following commands are based on this (otherwise just omit the `--user`).
Make sure to modify the .service to fit your needs, this being path of binary and config. Afterwards simply use the following commands.

To reload the services files and pick up your new .service file:
`systemctl --user daemon-reload`

To actually start the bot using systemd
`systemctl --user start eventbot`

To make it start on boot
`systemctl --user enable eventbot`

In the later case you'll probably also want to mark your user as lingering, otherwise systemd will just kill all your services the moment you logout.
In order to do this you'll have to do the following once:
`sudo loginctl enable-linger <user>`
Where of course `<user>` should be replaced with the user you're running the bot as.