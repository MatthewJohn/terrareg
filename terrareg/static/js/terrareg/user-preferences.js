
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

function getUserPreferences() {
    return {
        'show-beta-versions': getLocalStorageValue('show-beta-versions', false, Boolean),
        'show-unpublished-versions': getLocalStorageValue('show-unpublished-versions', false, Boolean)
    }
}

function userPreferencesModalShow() {
    let currentPreferences = getUserPreferences();
    $('#user-preferences-show-beta').prop('checked', currentPreferences["show-beta-versions"]);
    $('#user-preferences-show-unpublished').prop('checked', currentPreferences["show-unpublished-versions"]);

    $('#user-preferences-modal')[0].classList.add('is-active');
}

function userPreferencesModalSave() {
    localStorage.setItem('show-beta-versions', $('#user-preferences-show-beta').is(':checked'));
    localStorage.setItem('show-unpublished-versions', $('#user-preferences-show-unpublished').is(':checked'));
    userPreferencesModalClose();
}

function userPreferencesModalClose() {
    $('#user-preferences-modal')[0].classList.remove('is-active');
}