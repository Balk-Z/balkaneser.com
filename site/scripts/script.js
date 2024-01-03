//Check darmode preference
if (localStorage.getItem('color-theme') === 'dark' || (!('color-theme' in localStorage) && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
  document.documentElement.classList.add('dark');
} else {
  document.documentElement.classList.remove('dark')
}

//Progress bar
//Source: https://alligator.io/js/progress-bar-javascript-css-variables/
//Also handles shadow of navbar on scroll
const handleProgressbar = () => {
  let h = document.documentElement,
    b = document.body,
    st = "scrollTop",
    sh = "scrollHeight",
    progress = document.querySelector("#progress"),
    scroll;
  let scrollpos = window.scrollY;
  let header = document.getElementById("header");
  let navContent = document.getElementById("nav-content")

  //Refresh scroll % width
  scroll = (h[st] || b[st]) / ((h[sh] || b[sh]) - h.clientHeight) * 100;
  if (progress != null) {
    progress.style.setProperty("--scroll", scroll + "%");
  }

  //Navbar shadow
  if (header != null) {
    if (scrollpos > 10 || !navContent.classList.contains("hidden")) {
      header.classList.add("bg-white", "shadow", "dark:bg-gray-800");
      header.classList.remove("dark:bg-gray-900")
      navContent.classList.add("dark:bg-gray-800")
      navContent.classList.remove("dark:bg-gray-900")

    } else {
      header.classList.remove("bg-white", "shadow", "dark:bg-gray-800");
      header.classList.add("dark:bg-gray-900")
      navContent.classList.remove("dark:bg-gray-800")
      navContent.classList.add("dark:bg-gray-900")
    }
  }
}
document.addEventListener("scroll", handleProgressbar)

//Javascript to toggle the menu
const showHideNav = (e) => {
  let toggle = document.getElementById("nav-toggle")
  let navContent = document.getElementById("nav-content")
  let navDropdown = document.getElementById("dropdown")
  let scrollpos = window.scrollY;

  // If menu button is clicked
  if (toggle.contains(e.target)) {
    // If navContent is currently hidden
    if (navContent.classList.contains("hidden")) {
      navContent.classList.remove("hidden")
      navContent.classList.add("dark:bg-gray-800")
      navContent.classList.remove("dark:bg-gray-900")

      header.classList.add("bg-white", "shadow", "dark:bg-gray-800");
      header.classList.remove("dark:bg-gray-900")
    } else {
      navContent.classList.add("hidden")
      // only remove header classes if scroll allows
      if (scrollpos == 0) {
        header.classList.remove("bg-white", "shadow", "dark:bg-gray-800");
      }
    }
    // If clicked outside navbar
  } else if (!navContent.contains(e.target)) {
    // Ensure navDropdown is also closed if clicked outside navContent
    navDropdown.checked = false
    navContent.classList.add("hidden")
    if (scrollpos == 0) {
      header.classList.remove("bg-white", "shadow", "dark:bg-gray-800");
      navContent.classList.remove("dark:bg-gray-800")
      navContent.classList.add("dark:bg-gray-900")
    }
  }
  // Toggle inline navbar dropdown
  let dropdownMenu = document.getElementById("dropdownmenu")
  navDropdown.checked ? dropdownMenu.classList.remove("hidden") : dropdownMenu.classList.add("hidden")

}
document.addEventListener("click", showHideNav)

//Handles darkmode and button img switching
//From: https://flowbite.com/docs/customize/dark-mode/ documentation is awesome
function handleDarkmode() {
  var themeToggleDarkIcon = document.getElementById('theme-toggle-dark-icon');
  var themeToggleLightIcon = document.getElementById('theme-toggle-light-icon');

  // Change the icons inside the button based on previous settings
  if (localStorage.getItem('color-theme') === 'dark' || (!('color-theme' in localStorage) && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    themeToggleLightIcon.classList.remove('hidden');
  } else {
    themeToggleDarkIcon.classList.remove('hidden');
  }

  var themeToggleBtn = document.getElementById('theme-toggle');

  themeToggleBtn.addEventListener('click', function () {

    // toggle icons inside button
    themeToggleDarkIcon.classList.toggle('hidden');
    themeToggleLightIcon.classList.toggle('hidden');

    // if set via local storage previously
    if (localStorage.getItem('color-theme')) {
      if (localStorage.getItem('color-theme') === 'light') {
        document.documentElement.classList.add('dark');
        localStorage.setItem('color-theme', 'dark');
      } else {
        document.documentElement.classList.remove('dark');
        localStorage.setItem('color-theme', 'light');
      }

      // if NOT set via local storage previously
    } else {
      if (document.documentElement.classList.contains('dark')) {
        document.documentElement.classList.remove('dark');
        localStorage.setItem('color-theme', 'light');
      } else {
        document.documentElement.classList.add('dark');
        localStorage.setItem('color-theme', 'dark');
      }
    }

  });
}

//Handles highlighting of navbar element if applicable
function handleNavbarHighlighting() {
  let currentPath = window.location.pathname
  let navbarElements = document.getElementById("nav-list").getElementsByTagName("a")
  for (let i = 0; i < navbarElements.length; i++) {
    const element = navbarElements[i];
    if (element.getAttribute("href") == currentPath) {
      element.classList.add("text-gray-900", "font-bold")
      element.classList.remove("text-gray-600")
    } else {
      element.classList.remove("font-bold", "text-gray-900")
      element.classList.add("text-gray-600")
    }
  }
}

// Add navbar and footer from components to pages that contain placeholder elements
// all JS that access navbar/footer elements has to happen after the loop to avoid null objects
async function injectComponents() {
  for (const entry of ["navbar", "footer"]) {
    await fetch(`/components/${entry}.html`).then(r => r.text()).then(text => document.getElementById(`${entry}-placeholder`).outerHTML = text)
  }
  // highlight current navbar element
  handleNavbarHighlighting()
  handleDarkmode()
  // In case user refreshes mid page 
  if (window.scrollY > 0) {
    handleProgressbar()
  }

}
injectComponents()

function validateUserInputAndReact() {
  let username = document.getElementById("username");
  let password = document.getElementById("password");
  let loginBox = document.getElementById("loginBox");
  let error_obj = document.getElementById("err")
  let error_text = "Please enter information into both fields :)"

  if (username.value == "" || password.value == "") {
    let error_div = `<div id="err" class="text-center text-4xl text-red-600 pb-4">${error_text}</div>`
    let auth_error = new DOMParser().parseFromString(error_div, "text/html").body.firstChild
    if (error_obj == null) {
      loginBox.insertBefore(auth_error, loginBox.firstChild)
    } else {
      error_obj.innerHTML = error_text
    }
    username.classList.add("border-red-600", "border-4")
    password.classList.add("border-red-600", "border-4")
    username.value = "";
    password.value = "";
    return false
  } else {
    if (error_obj != null) {
      username.classList.remove("border-red-600", "border-4")
      password.classList.remove("border-red-600", "border-4")
      loginBox.removeChild(error_obj)
      if (error_obj.innerHTML == error_text) {
      alert("Good Job!")
      }
    }
    return true
  }
}

function btnlogin(e) {
  e.preventDefault();
  let error_text = "Authentication Error"
  let error_div = `<div id="err" class="text-center text-4xl text-red-600 pb-4">${error_text}</div>`

  if (validateUserInputAndReact()){
  let headers = new Headers();
  headers.set('Authorization', 'Basic ' + btoa(username.value + ":" + password.value));

  fetch("http://localhost:8080/login", {
    method: 'GET',
    headers: headers,
  }).then(response => {
    if (!response.ok) {
      throw Error();
    }
    

  }).catch(e => {
    let auth_error = new DOMParser().parseFromString(error_div, "text/html").body.firstChild
    if (document.getElementById("err") == null) {
      loginBox.insertBefore(auth_error, loginBox.firstChild)
    }
  });
  }

  username.value = "";
  password.value = "";
}

document.addEventListener("submit", btnlogin)