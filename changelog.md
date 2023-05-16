## Version 0.5.0
 - Add Terminal to containers
 - Add Create Container
 - Add support for importing Docker Compose
 - Fixed 2 bugs with the smart shield, that made it too strict
 - Fixed issues that prevented from login in with different hostnames
 - Added more infoon the shield when blocking someone
 - Fixed home background image
 - Fixed issue where the UI would have missing icon images
 - Improved icon fetching for apps by following redirections

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