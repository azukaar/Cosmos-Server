## Version 0.12.0
 - New Dashboard
 - New metrics gathering system
 - Integrated a new docker-less mode of functioning for networking
 - Added Button to force reset HTTPS cert in settings

## Version 0.11.3
 - Fix missing even subscriber on export

## Version 0.11.2
 - Improve Docker exports logs

## Version 0.11.1
 - fix issue exporting text user node

## Version 0.11.0
 - Disable support for X-FORWARDED-FOR incoming header (needs further testing)
 - Docker export feature for backups on every docker event
 - Compose Import feature now supports skipping creating existing resources
 - Compose Import now overwrite containers if they are differents
 - Added support for cosmos-persistent-env, to persist password when overwriting containers (useful for encrypted or password protected volumes, like databases use)
 - Fixed bug where import compose would try to revert a previously created volume when errors occurs
 - Terminal for import now has colours 
 - Fix a bug where ARM CPU would not be able to start Constellation

## Version 0.10.4
 - Encode OpenID .well-known to JSON
 - Fix incompatibility with other apps using .well-known
 - Secure the OpenID routes that missed the hardening
 - Added some logs

## Version 0.10.3
 - Add missing Constellation logs when creating certs
 - Ignore empty links in cosmos-compose

## Version 0.10.2
 - Fix port in host header

## Version 0.10.1
 - Fix an issue where Constellation is stuck if creating a new network is interrupted
 - Fix a logic issue with the whitelist inbound IPs

## Version 0.10.0
 - Added Constellation
 - DNS Challenge is now used for all certificates when enabled [breaking change]
 - Rework headers for better compatibility
 - Improve experience for non-admin users
 - Fix bug with redirect on logout
 - Added OverwriteHostHeader to routes to override the host header sent to the target app
 - Added WhitelistInboundIPs to routes to  filter incoming requests based on IP per URL

 > **Note: If you use the ARM (:latest-arm) you need to manually update to using the :latest tag instead**
 
## Version 0.9.20 - 0.9.21
 - Add option to disable CORS hardening (with empty value)

## Version 0.9.19
 - Add country whitelist option to geoblocker
 - No countries blocked by default anymore
 - Merged ARM and AMD into a single docker tag (latest)
 - Update to Debian 12
 - Fix issue with Contradictory scheme headers
 - Fix issue where non-admin users cant see Servapp on the homepage

## Version 0.9.18
 - Typo with x-forwarded-host

## Version 0.9.17
 - Upgraded to Lego 4.13.3 (support for Google Domain)
 - Add VerboseForwardHeader to URL Config to allow to transfer more sensitive header to target app
 - App DisableHeaderHardening to allow disabling header hardening for specific apps
 
## Version 0.9.16
 - Small redirection bug fix

## Version 0.9.15
 - Check background extension on upload is an image
 - Update Docker for security patch
 - Check redirect target is local
 - Improve OpenID client secret generation

## Version 0.9.14
 - Check network mode before pruning networks

## Version 0.9.13
 - Fix issue with duplicated ports in network tab of servapps (because it shows the IPV4 and the IPV6 ports)

## Version 0.9.12
 - Add integration to the `docker login` credentials store
 - Smart-shield now works with different budgets per routes, so that requests on a permissive route don't count as requests on a strict route
 - Fix an issue where users would never receive permanent bans from the shield

## Version 0.9.11
 - Add support for port ranges in cosmos-compose
 - Fix bug where multiple host port to the same container would override each other
 - Port display on Servapp tab was inverted
 - Fixed Network screen to support complex port mappings
 - Add support for protocol in cosmos-compose port exposing logic
 - Add support for relative bind path in docker-compose import
 - Fix environment vars and labels containing multiple equals (@jwr1)
 - Fix link to Other Setups page (@jwr1)

## Version 0.9.10
 - Never ban gateway ips
 - Prevent deleting networks if there's an error on disconnect
 - Disabling network pruning now also disables cleaning up Cosmos networks

## Version 0.9.9
 - Add new filters for routes based on method, query strings and headers (missing UI)

## Version 0.9.1 > 0.9.8
 - Fix subdomain logic for composed TLDs
 - Add option for custom wildcard domains
 - Fix domain depupe logic
 - Add import button in market
 - Update LEGO
 - Fix issue with hot-reloading between HTTP and HTTPS 
 - Fix loading bar in container overview page
 - Flush Etag cache on restart
 - Add timeout to icon fetching
 - Bootstrap containers when adding new routes to them
 - Remove headers from origin server to prevent duplicates
 - Add licence

## Version 0.9.0
 - Rewrote the entire HTTPS / DNS challenge system to be more robust and easier to use
   - Let's Encrypt Certificate is now saved in the config file
   - Cosmos will re-use previous certificate if renewal fails
   - Self-Signed certificate will now renew on expiry
   - If LE fails to renew, Cosmos will fallback to self-signed certificate
   - If LE fails to renew, Cosmos will display a warning on the home page
   - If certificate have more hostnames than required, Cosmos will not request a new certificate to prevent LE rate limiting issues
 - No more restart needed when changing config, adding route, installing apps, etc...
 - Change auto mapper to keep existing user definied ports
 - When using a subdomain as the main Cosmos domain, UseWildcardCertificate will now request the root domain instead of *.sub.domain.com
 - open id now supports multiple redirect uri (comma separated)
 - add manual restart button in config
 - New simpler Homepage style, with a toggle for expanded details homepage style in the config
 - add a button on the first setup screen to perform a clean install

## version 0.8.1 -> 0.8.10
 - Added new automatic Docker mapping feature (for people not using (sub)domains)
 - Added guardrails to prevent Let's Encrypt from failing to initialize when adding wrong domains
 - Add search bar on the marketplace
 - App store image size issue
 - Display more tags in the market
 - Fixed wrong x-forwarded-proto header 
 - Add installer option for hostname prefix/suffix
 - Fix minor issue with inconsistent password on market installer
 - Fixed issue where home page was https:// links on http only servers
 - Improved setup flow for setting up hostname and HTTPS
 - Fixed auto-update on ARM based CPU
 - Fix issue with email links
 - HideFromDashboard option on routes
 - Fix docker compose import issue with uppercase volumes

## Version 0.8.0
 - Custmizable homepage / theme colors
 - Auto-connect containers that have SERVAPP routes attached to them. aka. you do not need to "force secure" containers anymore
 - Manually create smaller docker subnets when using force secure / links to not hit IP range limit
 - Self-heal containers that have lost their network configurations
 - Stop showing Docker not connected when first loading status in new installs
 - Add a cosmos-icon label to containers to change the icon in the UI
 - Add privacy settings to external links
 - Force secure is now called "isolate network" to make it more clear, but does the same thing
 - allow iframes in the same subdomain as the app to fix wordpress compatibility

## Version 0.7.1 -> 0.7.10
 - Fix issue where multiple DBs get created at the setup
 - Add more special characters to be used for password validation
 - Add configurable default data path for binds
 - Remove Redirects from home page
 - Fix compat with non-HTTP protocol like WebDAV (for Nextcloud for example)
 - Fix regression with DNS wildcards certificates
 - Fix issue with the installer when changing both the labels and the volumes
 - Fix regression where DNS keys don't appear in the config page after being changed
 - Fix typo on "updating ServApp" message

## Version 0.7.0
 - Add Cosmos App Market!
 - Reforged the DNS CHallenge to be more user friendly. You can select your DNS provider in a list, and it will guide you through the process with the right fields to set (directly in the UI). No more env variables to set!
 - Fix issue with docker compose timeout healthcheck as string, inverted ports, and supports for uid:gid syntax in user
 - Fix for SELinux compatibility
 - Fix false-negative error message on login screen when SMTP is disabled

## Version 0.6.1 - 0.6.4
 - Workaround for Docker-compose race condition in Debian
 - Fix ARM based MongDB image for older ARM Devices
 - Fix issue with missing auth key with OpenID

## Version 0.6.0
- OpenID support!
- Add hostname check when adding new routes to Cosmos
- Add hostname check on new Install
- Fix missing save button for network mode

## Version 0.5.11
- Improve docker-compose import support for alternative syntaxes
- Improve docker service creation when using force secure label (fixes few containers not liking restarting too fast when created)
- Add toggle for using insecure HTTPS targets (fixes Unifi controller)

## Version 0.5.1 -> 0.5.10
- Add Wilcard certificates support
- Auto switch to Mongo 4 if CPU has no ADX
- Improve setup for certificates on new install
- Fix issue docker compose import labels and networks array
- Fix issue docker compose one-service syntax
- Fix issue with docker network mode not supporting hostname
- Fix an issue with the shield and the docker networking
- Fix issue with network namespace
- Fixed issue with a Docker bug preventing re-creating a container with a network mode as container (https://github.com/portainer/portainer/issues/2657)
- Silent error on favicon fetching
- Create Servapp step 1: make name / image required

## Version 0.5.0
 - Add Terminal to containers
 - Add "Create ServApp"
 - Add support for importing Docker Compose
 - Improved icon fetching 
 - Change Home background and style (especially fixing the awckward light theme)
 - Fixed 2 bugs with the smart shield, that made it too strict
 - Fixed issues that prevented from login in with different hostnames
 - Added more info on the shield when blocking someone
 - Fixed issue where the UI would have missing icon images
 - Fixed Homepage showing stopped containers
 - Fixed bug where you can't save changes on the URLs Screen

## Version 0.4.3
 - Fix for exposing routes from the details page

## Version 0.4.2
 -  Fix when using custom port and logging in (Isssue #10)

## Version 0.4.1
 - Fix small UI issues
 - Fix HTTP login

## Version 0.4.0
 - Protect server against direct IP access
 - Improvements to installer to make it more robust
 - Fix bug where you can't complete the setup if you don't have a database
 - When re-creating a container to edit it, restore the previous container if the edit is not succesful
 - Stop / Start / Restart / Remove / Kill containers
 - List / Delete / Create Volumes
 - List / Delete / Create  Networks
 - Container Logs Viewer
 - Edit Container Details and Docker Settings
 - Set Labels / Env variables on containers
 - (De)Attach networks to containers
 - (De)Attach volumes  to containers

## Version 0.3.1 -> 0.3.5
 - Fix UI issue with long name in home
 - Fix ARM docker image
 - Add more validation for Let's Encrypt
 - Prevent browser from auto-filling password in config page
 - Revert to HTTP when Let's Encrypt fails to initialize 

## Version 0.3.0
 - Implement 2 FA
 - Implement SMTP to Send Email (password reset / invites)
 - Add homepage
 - DNS challenge for letsencrypt
 - Set Max nb simulatneous connections per user 
 - Admin only routes (See in security tab)
 - Set Global Max nb simulatneous connections
 - Block based on geo-locations
 - Block common bots
 - Display nickname on invite page
 - Reset self-signed certificates when hostnames changes
 - Edit user emails
 - Show loading on user rows on actions

## Version 0.2.0
 - URL UI completely redone from scratch
 - Add new "Smart Shield" feature for easier protection without manual adjustments required
 - Add icons for self-hosted apps
 - Rewrite the restart function to allow the UI to gracefully wait for the server to restart
 - /login redirect now has query strings
 - prevent ports or network to scroll view
 - Fix URLs appearing on the wrong container because of nested names
 - Improve port display
 - Config API now reads the file directly to prevent overwritting changes between restarts
 - Warn user when there are config changes pending restart
 - Prevent login screen loop when being rate limited
 - Improve automatic hostname for new containers URLs
 - Fix minor bugs when host or prefix are false but values are set anyway
 - Edit should not reconnect bridge if force secure is true, for faster container restart
 - Improve network cleaning to prevent any issue with Docker Compose
 - Add Max Bandwith to routes to limit the amount of data that can be sent per seconds
 - Fix a bug where URLs target can't be edited if the container is in exited state
 - Fix bugs where the user would be editting the configuration on multiple tabs and end up in a bad state
 - Ensure route name is unique

## Version 0.1.16

 - Fix search
 - Fix bug where containers would lose their networks after being edited
 - Self-heal secure network configuration
 - Auto disconnect from orphan networks
 - Prevent bootstrapping from creating orphan networks
 - Monitor Docker and self-heal when docker daemon dies
 - Recreate lost secure networks (ex. when resetting Cosmos)



## Version 0.1.15 

 - Ports is now freetype, in case container does not expose any
 - Container picker now tries to pick the best port as default
 - Hostname now default to container name
 - Additional UI improvements