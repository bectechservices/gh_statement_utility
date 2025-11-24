(function () {
  var password = document.querySelector('.password');
  var passwordConfirmation = document.querySelector('.password_confirm');
  var helperText = {
    special: document.querySelector('.helper-text .special'),
    lowercase: document.querySelector('.helper-text .lowercase'),
    uppercase: document.querySelector('.helper-text .uppercase'),
    hasNumber: document.querySelector('.helper-text .hasNumber'),
    charLength: document.querySelector('.helper-text .length'),
    match: document.querySelector('.helper-text .match'),
  };
  const PASSWORD_LENGTH = window.scb_PASSWORD_LENGTH;
  var pattern = {
    charLength: function () {
      var characterCount = /^.{10,}/;
      if (characterCount.test(password.value)) {
        return true;
      }
    },
    lowercase: function () {
      var regex = /^(?=.*[a-z]).+$/; // Lowercase character pattern
      if (regex.test(password.value)) {
        return true;
      }
    },
    uppercase: function () {
      var regex = /^(?=.*[A-Z]).+$/; // Uppercase character pattern
      if (regex.test(password.value)) {
        return true;
      }
    },
    number: function () {
      var hasNumber = /^(?=.*[\d]).+$/;
      if (hasNumber.test(password.value)) {
        return true;
      }
    },
    special: function () {
      var regex = /([-+=_!@#$%^&*.,;:'\"<>/?`~\[\]\(\)\{\}\\\|\s])/; // Special character or number pattern
      if (regex.test(password.value)) {
        return true;
      }
    }
  };

  password.addEventListener('keyup', function () {
    patternTest(pattern.charLength(), helperText.charLength);
    patternTest(pattern.lowercase(), helperText.lowercase);
    patternTest(pattern.uppercase(), helperText.uppercase);
    patternTest(pattern.number(), helperText.hasNumber);
    patternTest(pattern.special(), helperText.special);


    if (password.value === passwordConfirmation.value && password.value.length !== 0 && passwordConfirmation.value.length !== 0) {
      addClass(helperText.match, 'valid');
    } else {
      removeClass(helperText.match, 'valid');
    }
    if (
      hasClass(helperText.charLength, 'valid') &&
      hasClass(helperText.lowercase, 'valid') &&
      hasClass(helperText.uppercase, 'valid') &&
      hasClass(helperText.special, 'valid') &&
      hasClass(helperText.hasNumber, 'valid')
    ) {
      addClass(password.parentElement, 'valid');
    } else {
      removeClass(password.parentElement, 'valid');
    }
  });
  passwordConfirmation.addEventListener('keyup', function () {
    if (password.value === passwordConfirmation.value && password.value.length !== 0 && passwordConfirmation.value.length !== 0) {
      addClass(helperText.match, 'valid');
    } else {
      removeClass(helperText.match, 'valid');
    }
    // Check that all requirements are valid
    if (
      hasClass(helperText.match, 'valid')
    ) {
      addClass(passwordConfirmation.parentElement, 'valid');
    } else {
      removeClass(passwordConfirmation.parentElement, 'valid');
    }
  });

  function patternTest(pattern, response) {
    if (pattern) {
      addClass(response, 'valid');
    } else {
      removeClass(response, 'valid');
    }
  }

  function addClass(el, className) {
    if (el.classList) {
      el.classList.add(className);
    } else {
      el.className += ' ' + className;
    }
  }

  function removeClass(el, className) {
    if (el.classList)
      el.classList.remove(className);
    else
      el.className = el.className.replace(new RegExp('(^|\\b)' + className.split(' ').join('|') + '(\\b|$)', 'gi'), ' ');
  }

  function hasClass(el, className) {
    if (el.classList) {
      return el.classList.contains(className);
    } else {
      new RegExp('(^| )' + className + '( |$)', 'gi').test(el.className);
    }
  }
})();

const form = document.querySelector('.authForm');
form.addEventListener('submit', function (e) {
  e.preventDefault();
  const isValid = function (className) {
    return document.querySelector(`.helper-text .${className}`).classList.contains('valid')
  };
  if (['special', 'lowercase', 'uppercase', 'hasNumber', 'length', 'match'].every(isValid)) {
    form.submit();
  }
})