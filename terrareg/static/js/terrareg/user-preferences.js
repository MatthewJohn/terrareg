function getLocalStorageValue(key, defaultValue, type) {
    let localStorageValue = localStorage.getItem(key);
    let value = localStorageValue === null ? defaultValue : localStorageValue;

    if (type === Boolean) {
        if (value == "false") {
            value = false;
        } else if (value == "true") {
            value = true;
        }
    }
    return value;
}

async function getUserPreferences() {
    let config = await getConfig();
    return {
        'show-beta-versions': getLocalStorageValue('show-beta-versions', false, Boolean),
        'show-unpublished-versions': getLocalStorageValue('show-unpublished-versions', false, Boolean),
        'theme': $.cookie("theme") || 'default',
        'terraform-compatibility-version': getLocalStorageValue('terraform-compatibility-version', '', String),
        'ui-details-view': getLocalStorageValue('ui-details-view', config.DEFAULT_UI_DETAILS_VIEW, String),
    }
}

function setUiDetailsView(newValue) {
    localStorage.setItem("ui-details-view", newValue);
}

function userPreferencesUpdateTerraformCompatibilityVersion(newVersion) {
    localStorage.setItem('terraform-compatibility-version', newVersion);
}

async function userPreferencesModalShow() {
    let currentPreferences = await getUserPreferences();
    $('#user-preferences-show-beta').prop('checked', currentPreferences["show-beta-versions"]);
    $('#user-preferences-show-unpublished').prop('checked', currentPreferences["show-unpublished-versions"]);
    $('#user-preferences-theme').val(currentPreferences["theme"]);
    $('#user-preferences-terraform-compatibility-version').val(currentPreferences["terraform-compatibility-version"]);

    $('#user-preferences-modal')[0].classList.add('is-active');
}

function userPreferencesModalSave() {
    localStorage.setItem('show-beta-versions', $('#user-preferences-show-beta').is(':checked'));
    localStorage.setItem('show-unpublished-versions', $('#user-preferences-show-unpublished').is(':checked'));
    localStorage.setItem('terraform-compatibility-version', $('#user-preferences-terraform-compatibility-version').val());
    localStorage.setItem('theme', $('#user-preferences-theme').val());
    $.cookie("theme", $('#user-preferences-theme').val(), {path: "/", expires: 365});
    userPreferencesModalClose();

    location.reload();
}

function userPreferencesModalClose() {
    $('#user-preferences-modal')[0].classList.remove('is-active');
}