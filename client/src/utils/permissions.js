// Permission constants — must match src/utils/types.go
export const PERM_ADMIN_READ = 1;
export const PERM_ADMIN = 2;
export const PERM_USERS_READ = 10;
export const PERM_USERS = 11;
export const PERM_RESOURCES_READ = 20;
export const PERM_RESOURCES = 21;
export const PERM_CONFIGURATION_READ = 30;
export const PERM_CONFIGURATION = 31;
export const PERM_CREDENTIALS_READ = 40;
export const PERM_LOGIN = 100;
export const PERM_LOGIN_WEAK = 101;

// Global: permissions that require sudo for ANY role
// Must match SudoPermissions in types.go
export const SUDO_PERMISSIONS = [PERM_ADMIN, PERM_USERS, PERM_RESOURCES, PERM_CONFIGURATION, PERM_CREDENTIALS_READ];
