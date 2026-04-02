/* ===================================================================
   YMMO — main.js
   =================================================================== */

// --- Navbar scroll effect ---
(function () {
  const navbar = document.querySelector('.navbar');
  if (!navbar) return;
  const onScroll = () => navbar.classList.toggle('scrolled', window.scrollY > 40);
  window.addEventListener('scroll', onScroll, { passive: true });
  onScroll();
})();

// --- Toggle user dropdown ---
function toggleUserMenu() {
  const dropdown = document.getElementById('userDropdown');
  if (!dropdown) return;
  const isOpen = dropdown.classList.toggle('open');
  const userBtn = document.querySelector('.nav-user');
  if (userBtn) {
    userBtn.setAttribute('aria-expanded', isOpen);
  }
}

// Close on outside click
document.addEventListener('click', function (e) {
  const dropdown = document.getElementById('userDropdown');
  if (!dropdown) return;
  if (!e.target.closest('.nav-user')) {
    dropdown.classList.remove('open');
    const userBtn = document.querySelector('.nav-user');
    if (userBtn) userBtn.setAttribute('aria-expanded', 'false');
  }
});

// --- Mobile menu ---
function toggleMobileMenu() {
  const mobileNav = document.getElementById('mobileNav');
  const hamburger = document.querySelector('.hamburger');
  if (!mobileNav) return;
  const isOpen = mobileNav.classList.toggle('open');
  if (hamburger) hamburger.setAttribute('aria-expanded', isOpen);
  mobileNav.setAttribute('aria-hidden', !isOpen);
  document.body.style.overflow = isOpen ? 'hidden' : '';
}

// Close mobile nav when a link is clicked
document.addEventListener('DOMContentLoaded', function () {
  const mobileLinks = document.querySelectorAll('.mobile-nav-link');
  mobileLinks.forEach(function (link) {
    link.addEventListener('click', function () {
      const mobileNav = document.getElementById('mobileNav');
      if (mobileNav) mobileNav.classList.remove('open');
      if (mobileNav) mobileNav.setAttribute('aria-hidden', 'true');
      const hamburger = document.querySelector('.hamburger');
      if (hamburger) hamburger.setAttribute('aria-expanded', 'false');
      document.body.style.overflow = '';
    });
  });
});

// --- Toggle password visibility ---
function togglePwd(inputId, btn) {
  const input = document.getElementById(inputId);
  if (!input) return;
  const isHidden = input.type === 'password';
  input.type = isHidden ? 'text' : 'password';
  
  // Swap icon if using FontAwesome or similar, otherwise fallback to opacity
  const icon = btn.querySelector('i');
  if (icon) {
    icon.className = isHidden ? 'fas fa-eye-slash' : 'fas fa-eye';
  } else {
    btn.style.opacity = isHidden ? '1' : '0.5';
  }
}

// --- Password strength indicator ---
function checkStrength(password, barId) {
  const bar = document.getElementById(barId);
  if (!bar) return;
  bar.className = 'strength-fill';
  if (!password) return;
  let score = 0;
  if (password.length >= 8) score++;
  if (/[A-Z]/.test(password)) score++;
  if (/[0-9]/.test(password)) score++;
  if (/[^A-Za-z0-9]/.test(password)) score++;
  const classes = ['', 's1', 's2', 's3', 's4'];
  bar.classList.add(classes[score]);
}

// --- Auto-dismiss flash messages ---
document.addEventListener('DOMContentLoaded', function () {
  const flashes = document.querySelectorAll('.flash');
  flashes.forEach(function (flash) {
    var delay = parseInt(flash.dataset.delay || '4000', 10);
    var timer = setTimeout(function () { dismissFlash(flash); }, delay);
    var closeBtn = flash.querySelector('.flash-close');
    if (closeBtn) {
      closeBtn.addEventListener('click', function () {
        clearTimeout(timer);
        dismissFlash(flash);
      });
    }
  });
});

function dismissFlash(el) {
  el.style.transition = 'opacity 0.3s, transform 0.3s';
  el.style.opacity = '0';
  el.style.transform = 'translateX(110%)';
  setTimeout(function () { el.remove(); }, 320);
}

// --- Gallery lightbox ---
(function () {
  const mainImg = document.getElementById('galleryMain');
  const lightbox = document.getElementById('lightbox');
  const lightboxImg = document.getElementById('lightboxImg');

  if (mainImg && lightbox && lightboxImg) {
    mainImg.addEventListener('click', function () {
      lightboxImg.src = mainImg.src;
      lightbox.classList.add('open');
      lightbox.setAttribute('aria-hidden', 'false');
      // Ensure the lightbox is focusable to trap keyboard navigation
      lightbox.setAttribute('tabindex', '-1');
      lightbox.focus();
    });
    lightbox.addEventListener('click', function (e) {
      if (e.target === lightbox || e.target.classList.contains('lightbox-close')) {
        lightbox.classList.remove('open');
        lightbox.setAttribute('aria-hidden', 'true');
      }
    });
    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape') {
        lightbox.classList.remove('open');
        lightbox.setAttribute('aria-hidden', 'true');
      }
    });
  }

  // Thumbnails
  const thumbs = document.querySelectorAll('.gallery-thumb');
  thumbs.forEach(function (thumb) {
    thumb.addEventListener('click', function () {
      if (mainImg) mainImg.src = thumb.querySelector('img').src;
    });
  });
})();

// --- Favorite toggle (AJAX) ---
function toggleFavorite(propertyId) {
  fetch('/favoris/' + propertyId, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    credentials: 'same-origin'
  })
    .then(function (res) {
      if (res.redirected) {
        window.location.href = res.url;
        return;
      }
      if (!res.ok) return;
      var btn = document.getElementById('fav-btn-' + propertyId);
      if (btn) {
        btn.classList.toggle('fav-active');
        const icon = btn.querySelector('i');
        if (icon) {
          icon.classList.toggle('fas'); // Solid (active)
          icon.classList.toggle('far'); // Regular/Outline (inactive)
        }
      }
    })
    .catch(function () {});
}

// --- Demo account fill (login page) ---
function fillDemo(email, password) {
  var emailInput    = document.getElementById('email');
  var passwordInput = document.getElementById('password');
  if (emailInput)    emailInput.value    = email;
  if (passwordInput) passwordInput.value = password;
}

// --- Confirm on destructive forms ---
document.addEventListener('DOMContentLoaded', function () {
  document.querySelectorAll('form[data-confirm]').forEach(function (form) {
    form.addEventListener('submit', function (e) {
      if (!confirm(form.dataset.confirm)) e.preventDefault();
    });
  });
});

// --- Form Validation & Loading States ---
document.addEventListener('DOMContentLoaded', function () {
  const forms = document.querySelectorAll('form:not([data-no-validate])');
  forms.forEach(function (form) {
    form.addEventListener('submit', function (e) {
      // HTML5 Form Validation
      if (!form.checkValidity()) {
        e.preventDefault();
        e.stopPropagation();
        form.classList.add('was-validated');
        return;
      }
      
      // Prevent double submission and show loading spinner
      const submitBtn = form.querySelector('button[type="submit"]');
      if (submitBtn && !submitBtn.dataset.loading) {
        submitBtn.dataset.loading = "true";
        const originalWidth = submitBtn.offsetWidth;
        submitBtn.style.width = originalWidth + 'px'; // Maintain width
        submitBtn.innerHTML = '<i class="fas fa-circle-notch fa-spin"></i> Traitement...';
        submitBtn.disabled = true;
      }
    });
    
    // Live validation feedback on input after first failed attempt
    const inputs = form.querySelectorAll('input, select, textarea');
    inputs.forEach(input => {
      input.addEventListener('input', () => {
        if (form.classList.contains('was-validated')) {
          input.classList.toggle('is-invalid', !input.checkValidity());
          input.classList.toggle('is-valid', input.checkValidity());
        }
      });
    });
  });
});

// --- Search form: city quick-select from city cards ---
document.addEventListener('DOMContentLoaded', function () {
  document.querySelectorAll('.city-card[data-city]').forEach(function (card) {
    card.addEventListener('click', function () {
      var cityField = document.getElementById('search-city');
      if (cityField) {
        cityField.value = card.dataset.city;
        var form = cityField.closest('form');
        if (form) form.submit();
      }
    });
  });
});

// --- Scroll Reveal Animations ---
document.addEventListener('DOMContentLoaded', function () {
  const observerOptions = {
    root: null,
    rootMargin: '0px',
    threshold: 0.1
  };

  const observer = new IntersectionObserver(function(entries, observer) {
    entries.forEach(function(entry) {
      if (entry.isIntersecting) {
        entry.target.classList.add('reveal-visible');
        observer.unobserve(entry.target);
      }
    });
  }, observerOptions);

  document.querySelectorAll('.reveal').forEach(function(el) {
    observer.observe(el);
  });
});
