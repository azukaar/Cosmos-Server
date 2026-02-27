## Version 0.21.5
 - Fix issue with nodes having 2 public hostnames

## Version 0.21.4
 - Remove tunnel check when Constellation is off 

## Version 0.21.3
 - fix Bug with table width
 - Add move to top/bottom
 - Add chmod Restic

## Version 0.21.2
 - Publish translations 
 
## Version 0.21.1
 - Fix regression with Constellation creation 

## Version 0.21.0
 - UI refresh for most pages
 - Reworked mount managemenet logic for flexibility
 - Added "skip clean URL" option for apps that have invalid URLs like Synology
 - Support for tempFS
 - Improve support for 0.0.0.0 routes
 - Reworked the setting page for better clarity
 - Added support for HTTP2 Cleartext (H2C)
 - Added advanced DNS options (challenge related, thanks @moham96)
 - Added trusted proxy option for X-Forwarded-For (Thanks @InterN0te)
 - Reduce SmartShield false positive on the server panel by having two level of strictness (UI/panel vs. login)
 - Rclone rework, no more sub processes (increased performance and reliability)
 - Added the ability to manage Restic locks from the UI
 - Reworked Constellation cluster synchronisation completely. Now there are no more "Master" server and each server is equally capable.
 - Any server can be used as DNS (add redundancy too)
 - Any server can tunnel another server's URL in any direction
 - Constellation can now use arbitrary IP range
 - Server can now modify their setup without having to recreate the network
 - Fully reworked licencing system should reduce friction when moving / resetting servers

## Version 0.20.2
 - Fix regression with header hardening

## Version 0.20.1
 - Fix Avahi not working issue
 - Fix issue with resetting mem limit to 0

## Version 0.20.0
 - Added Samba for both remote storage and serve share
 - Added .env file upload when uploading compose files
 - You can now upload a custom icon on URLs
 - Fixed issue with non-admin users not seeing custom container icon
 - Fixed issue with non-admin users seeing stopped containers URLs on the dashboard
 - Improved handling of Docker login for private docker images
 - Support for hardware constraint (CPU/Mem/...) on Docker containers
 - Installer now installs fuse3 for Rclone
 - Fix issue with Rclone cache duration
 - Enable CORS passthrought when hardening is OFF
 - Added autocomplete on login for password managers
 - Updated Lego to v4.31.0
 - Fix crash when the authentication database is un-openable
 - Make VPN less verbose in logs
 - Redirect URLs now show on the dashboard (use the "hide from dashboard" option on URLs to hide them yourself)

## Version 0.19.1
 - Updated to Mongo 8 by default
 - Fixed update error for docker-container installations of Cosmos

## Version 0.19.0
 - Constellation allows nodes to see and ping each others
 - Constellation now has a firewall!
 - Constellation now has exit nodes
 - Constellation now automatically resolve the mesh before connecting
 - Improve docker image cleanup efficiency
 - Improve support for container network modes in import/export
 - Fixed the annoying "user unauthenticated" error when opening the homepage after the admin token expired
 - Fixed issue with exporting hostname when it would be incompatible to re-importing it
 - Updating network mode now also updates the network-mode label
 - Default storage path is now /cosmos-storage instead of /usr
 - Fixed bug where you cant delete the same device twice from Constellation
 - Export all containers do not export puppet containers anymore
 - container edits now respect the force network label
 - New licence field in the UI, more comprehensible
 - Licence change: Licence accomodates 20 users, 200 constellation devices but also TWO cosmos server (as long as they are in the same constellation. Do not use the licence twice, instead let constellation create a second licence)

## Version 0.18.4
 - Fix issue with DB credentials dissapearing
 - Remove expired discount

## Version 0.18.3
 - Fix Cosmos not creating repository for internal config

## Version 0.18.2
 - Add tooltip about admin in the market

## Version 0.18.1
 - Fix missing signing key generation

## Version 0.18.0
 - UI to backup and restore containers/folders/volumes using Restic
 - Implements sudo mode - your normal token last longer, but you need to "sudo" to do admin tasks
 - Re-Implements the SSO using openID internally - fixes issue where you need to re-loging when app are on different domains (because of browser cookies limitations)
 - Implements local HTTPS Certificate Authority, to locally trust self-signed certificates on devices
 - Added new folder button to file picker
 - Cosmos now waits for CRON jobs to be over before restarting the server
 - Fixed bug with RClone storage duplication in the UI
 - Implements hybrid HTTPS with public and self-signed certificates switched on the fly
 - OpenID now returns more info in case of errors when Cosmos is in debug mode
 - Localizations improvements (Thanks @madejackson)
 - Improved local IP detection (Thanks @r41d)
 - Updated LEGO to 4.21.0
 - Largely improved the experience of non-admin users (extra errors should all be gone)
 - Fixed file picker prefix issue in docker container
 - Added OpenID IDTokenSigningAlgValuesSupported
 - Added protocol in openid discovery endpoint
 - Fix RClone not starting (hopefully)
 - Added traditional Chinese translation
 - Avahi now ignores virtual interfaces
 - Fixed bug preventing the local mDNS broadcaster from publishing over 17 entries
 - Fixed bug with restarting slave Constellation node's Nebula process

## Version 0.17.7
 - Fix error code on login screen

## Version 0.17.6
 - Can now chose between plain or formatted logs
 - Added "force update" button in service mode
 - Fixed issues related to fstab management
 
## Version 0.17.5
 - Fixed issue with TCP proxy timeouts

## Version 0.17.4
 - Hide update button in container
 - Fix issue with allowHTTPLocal and the TCP Proxy

## Version 0.17.3
 - fix race condition with the monitoring

## Version 0.17.2
 - Fix RClone false error
  
## Version 0.17.1
 - Increase user limit to 19
 - Add missing translations

## Version 0.17.0
 - Added RClone integration to manage and mount remote storage
 - Added network shares for SFTP, NFS, WebDAV, and S3, with support for remote storages and smart-shield integration
 - Added zip packages for docker-less distribution
 - Added TCP/UDP socket proxying
 - TCP/UDP proxying includes smart-shield protections, constellation support, geoblocking, and monitoring
 - Added terminal shortcut on the top right header
 - Added restart server and restart cosmos button on the top right header
 - Added log file in the config folder and a download button in the config page
 - Fixed bug where lighthouses would not consider the home server as lighthouse in Constellation
 - Improved terminal with better UI and keep alive in the Websocket for  (to prevent timeouts during long operations)
 - Fixed bug with duplicated CORS headers
 - Disabled CORS for routes that have hardening disabled
 - Improve logs screen for containers (better colors, fix scrolling, auto-refresh, ...)
 - Fix bug with missing post-install instructs on service creation
 - Added missing geo block events in monitoring
 - Added ExtraHeader to route config to add custom headers to the request
 - Improved accessiblity of the menu for screen readers
 - Formatter now creates GPT partition tables (instead of MBR, which has a 2TB limit)
 - Update to Go 1.23.2
 - Fix 2-parity on Snapraid
 - Fix mount/unmount request false error
 - Added safeguard to prevent Docker from destroying stack containers hostnames
 - Added hostname to some events for visibility (Thanks @InterN0te)
 - Added missing content type in OAuth (Thanks @RaidMax)

## Version 0.16.3
 - Local domains now produce services instead of CNAME for better compatibility
 - DNS Lookup is now a warning
 - DNS Lookup ignores local domains

## Version 0.16.2
 - Only propose cosmos.local as default to setup using local domains

## Version 0.16.1
 - Added DNSChallengeResolvers config to allow using custom DNS resolvers for the DNS challenge

## Version 0.16.0
 - Multilanguage support (Thanks @madejackson)
 - Added automatic mDNS publishing for local network
 - Improve offline mode with Constellation
 - Add automatic sync of Constellation nodes
 - Constellation is now paid
 - Nodes in a constellation can now auto-sync credentials
 - Improve DNS Challenge with smarter resolution for faster and more reliable results (especially when using local nameservers)
 - Fix issues where it was impossible to login with insecure local IPs
 - Better suppoer for container/service network_mode when importing compose
 - Default networks to 16 Ips instead of 8
 - Further improving the docker-compose import to mimic naming and hostnaming convention
 - Added hostname stickiness to compose network namespaces
 - Added depends_on conditions to compose import
 - Fixed issues with container's monitoring when name contains a dot (Thanks @BearTS)
 - Added email on succesful login  (Thanks @BearTS)
 - Add support for runtime (Thanks @ryan-schubert)
 - Revamped the header and sidebar a little
 - Improve Docker VM detection
 - Fix a small UI bug with the constellation tab where UI falls behind
 - Now supports multiple wildcards at the same time for the DNS challenge

## Version 0.15.7
 - Added "Allow insecure local connection" for HTTP ip:port access in the same network
 - Fix issue where Cosmos request IP based certs to LE if setup
 - Added a "duplicate route" button to URL management

## Version 0.15.6
 - accept any insecure TLS when option is on

## Version 0.15.5
 - Use a different IP scheme for containers

## Version 0.15.4
 - Added SpoofHostname to hack apps who hate reverse proxies
 - Fix forward headers once and for all
 - Fix inverted port setup in the create Servapp form (#232)
 - Fix the device field in the setup screen (#237)
 - Fix the device field in the create Servapp form
 - Fix bug where non-admin users try to show the cron job widget
 - Hide the scheduler/storage tab from non-admin users
 - Hide DNS provider env var from non-admin users
 - Fixed DB file permission issue (Thanks @george-radu-cs)
 - Improved setup screen performance (Thanks @davis4acca)

## Version 0.15.3
 - Add support for sysctls as array
 - Fix temperature appearing in the disk usage widget

## Version 0.15.2
 - Fix an issue with DB creations

## Version 0.15.1
 - Add a toggle for search engine indexing
 - Fix an issue with the TCP proxy and ports already bound
 - Improve subpath handling

## Version 0.15.0
 - Added Disk management (Format, mount, SMART, etc...)
 - Added MergerFS support and configurator
 - Added SnapRAID support and configurator
 - Rewrote the internal CRON scheduler to be more robust
 - Added support for custom CRON jobs
 - Added job scheduler management, with manual run, logs, cancellation, ...
 - Added new Terminal (with full bash support, including things like VIM)
 - Overwrite all docker networks size to prevent Cosmos from running out of IP addresses
 - Added optional subnet input to the network creation
 - Fix issue with Sysctl not being applied
 - Fixed RAM issues
 - Rewrite network pruning to prevent Docker from deleting networks attached to stopped containers
 - Restore static bundle loading to fix issue with some browsers
 - Fix issue on Macvlan creation
 - Rewrite SPA handler for more robustness
 - Added Robots.txt
 - Added "restart" as action for alerts
 - Make monitoring more reliant in case of issue
 - Added a memory profiler when debug mode is on (/cosmos/debug/pprof)
 - Fix a crash when adding a protocol to a host
 - Update Docker and LEGO (with a dozen new DNS providers supported)
 - Added optionals vars to DNS challenge setup (like timeout)
 - Added a check on hostname to prevent protocols
 - Added hint to TCP proxying
 - Fix issue with favicon retrieval post-migration to host mode

## Version 0.14.6
 - Fix custom back-up folder logic

## Version 0.14.5
 - Fix an issue with the whitelisting form of URLs

## Version 0.14.4
 - Fix issue with the volumes going read only

## Version 0.14.3
 - Added app count to the marketplace

## Version 0.14.2
 - Puppet migration for non-host mode
 - Add config UI for puppet mode

## Version 0.14.0
 - Cosmos is now fully functional dockerless
 - Reworked Cosmos Compose for better compatibility with docker-compose.yml files
 - Added a "compose" tab to edit containers in text mode
 - Moved critical data (credentials and VPN) out of the database, to keep Cosmos online in case of incidents
 - Added an auto .zip backup mechanism
 - Added a syntax highlighter to the compose editor
 - New Database "puppet" mode that allows Cosmos to manage the database for you
 - Improved network IP resolution for containers, including supporting any network mode
 - Added support for markets and template directly with docker-compose.yml files
 - Add whitelist and constellation restriction options to the admin panel 
 - Force low RAM usage on the MongoDB container (we don't need much!)
 - Removed all sort of container bootstrapping (much faster boot)
 - Added image clean up
 - Replaced network clean up by vanilla docker prune
 - Fix issue with removing IP whitelist
 - Add UI to create MCVlan networks
 - Add a log file in config folder for the selfupdater
 - Add a migration script to host mode
 - UI optimizations (thanks @Kawanaao)
 - Add duplicate filter on store listing
 - Fixed an issue where container picker would select 'null' as container
 - Fix bug where Enabled checkbox was broken after a search
 - remove mac address when switching to host mode

## Version 0.13.2
 - Fix display issue with fault network configurations

## Version 0.13.1
 - Fix a security issue with token (thansk @vncloudsco)

## Version 0.13.0
 - Display container stacks as a group in the UI
 - New Delete modal to delete services entirely
 - Upload custom icons to containers
 - improve backup file, by splitting cosmos out to a separate docker-compose.yml file
 - Cosmos-networks now have specific names instead for generic names
 - Fix issue where search bar reset when deleting volume/network
 - Fix breadcrumbs in subpaths
 - Remove graphs from non-admin UI to prevent errors
 - Rewrite the overwriting container logic to fix race conditions
 - Edit container user and devices from UI
 - Fix bug where Cosmos Constellation's UDP ports by a TCP one
 - Fix a bug with URL screen, where you can't delete a URL when there is a search
 - Fix issue where negative network rate are reported
 - Support array command and single device in docker-compose import
 - Add default alerts... by default (was missing from the default config)
 - disable few features liks Constellation, Backup and Monitoring when in install mode to reduce logs and prevent issues with the DB

## Version 0.12.6
 - Fix a security issue with cross-domain APIs availability

## Version 0.12.5
 - Added index on event date for faster query

## Version 0.12.4
 - Fix crash with metrics not seeing any network interface

## Version 0.12.3
 - Performance update for metrics saving

## Version 0.12.2
 - Fix XSS vulnerability in the redirect function (thanks @catmandx)

## Version 0.12.1
 - Fix a crash that would occasionally happen since 0.12 the DB is down

## Version 0.12.0
 - New real time persisting and optimized metrics monitoring system (RAM, CPU, Network, disk, requests, errors, etc...)
 - New Dashboard with graphs for metrics, including graphs in many screens such as home, routes and servapps
 - New customizable alerts system based on metrics in real time, with included preset for anti-crypto mining and anti memory leak 
 - New events manager (improved logs with requests and advanced search)
 - New notification system
 - Added Marketplace UI to edit sources, with new display of 3rd party sources
 - Added a notification when updating a container, renewing certs, etc...
 - Certificates now renew sooner to avoid Let's Encrypt sending emails about expiring certificates
 - Added option to disable routes without deleting them
 - Improved icon loading speed, and added proper placeholder
 - Marketplace now fetch faster (removed the domain indirection to directly fetch from github)
 - Integrated a new docker-less mode of functioning for networking
 - Added a dangerous IP detector that stops sending HTTP response to IPs that are abusing various shields features
 - Added CORS headers to openID endpoints
 - Added a button in the servapp page to easily download the docker backup
 - Added Button to force reset HTTPS cert in settings
 - Added lazyloading to URL and Servapp pages images
 - Fixed annoying marketplace screenshot bug (you know what I'm talking about!)
 - New color slider with reset buttons
 - Redirect static folder to host if possible
 - New Homescreen look
 - Fixed blinking modals issues
 - Add AutoFocus on Token field for 2FA Authentication (thanks @InterN0te)
 - Allow Insecure TLS like self-signed certificate for SMTP server (thanks @InterN0te)
 - Improve display of icons [fixes #121]
 - Refactored Mongo connection code [fixes #111]
 - Forward simultaneously TCP and UDP [fixes #122]

## Version 0.11.3
 - Fix missing event subscriber on export

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
