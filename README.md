# service-notifications

A tool that creates slack channels for services in planning center and adds people who are assigned to the plan. This is to make it easy to communicate with people assigned to a plan, either automatically via the API included with tool, or manually in Slack. I wrote this tool to send notifications when a slide in ProPresenter is clicked, using the https://github.com/GRMrGecko/midi-request-trigger MIDI bridge.

## Install

You can install by building.

### Building

Building should be as simple as running:

```bash
go build
```

### Running as a service

You are likely going to want to run the tool as a service to ensure it runs at boot and restarts in case of failures. Below is an example service config file you can place in `/etc/systemd/system/service-notifications.service` on a linux system to run as a service if you install the binary in `/usr/local/bin/`.

```systemd
[Unit]
Description=Service Notifications
After=network.target
StartLimitIntervalSec=500
StartLimitBurst=5

[Service]
ExecStart=/usr/local/bin/service-notifications
ExecReload=/bin/kill -s HUP $MAINPID
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

Once the service file is installed, you can run the following to start it:

```bash
systemctl daemon-reload
systemctl start service-notifications.service
```

On MacOS, you can setup a Launch Agent in `~/Library/LaunchAgents/com.mrgeckosmedia.service-notifications.plist` as follows:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.mrgeckosmedia.service-notifications</string>
	<key>ProgramArguments</key>
	<array>
		<string>/path/to/bin/service-notifications</string>
        <string>-c</string>
        <string>/path/to/config.yaml</string>
	</array>
	<key>KeepAlive</key>
	<dict>
		<key>Crashed</key>
		<true/>
		<key>SuccessfulExit</key>
		<false/>
	</dict>
	<key>RunAtLoad</key>
	<true/>
    <key>OnDemand</key>
    <false/>
</dict>
</plist>

```

Start with:
```bash
launchctl load ~/Library/LaunchAgents/com.mrgeckosmedia.service-notifications.plist
```

Check status with:
```bash
launchctl list com.mrgeckosmedia.service-notifications
```

Stop with:
```bash
launchctl unload ~/Library/LaunchAgents/com.mrgeckosmedia.service-notifications.plist
```

## Cron job

The idea is to setup cron jobs to update data/create the slack channels on a particular day. The following is an example of what I would use.

```crontab
0 6 * * 3 /path/to/bin/service-notifications --update
```

## Config

The default configuration paths are:

- `./config.yaml` - A file in the current working directory.
- `~/.config/service-notifications/config.yaml` - A file in your home directory's config path.
- `/etc/service-notifications/config.yaml` - A file in the etc config folder.

### Basic config

Get Slack API token by creating an app at https://api.slack.com/apps then go to "Install App" to get the token.

Get Planning Center API secrets at https://api.planningcenteronline.com/oauth/applications by creating a personal access token.

You can get a slack user ID by viewing the profile and under the 3 dot menu choose Copy member ID.

```yaml
---
database:
    debug: true

planning_center:
    app_id: PC_APP_ID
    secret: PC_SECRET

slack:
    api_token: SLACK_API_TOKEN
    create_from_weekday: 3
    default_conversation: SLACK_UID
	sticky_users:
        - SLACK_UID

```