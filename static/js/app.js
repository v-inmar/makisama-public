// Make sure dom loaded before executing any scripts
document.addEventListener("DOMContentLoaded", () => {
  triggerFlashModal();
  clearFlashModalOnHidden();

  formatDateTimeRelative();
  formatDateTimeDefault();
  formatDateDefault();
  // cleanPendingFeedbackViewModal();

  // formatDateTimeNormal("datetime-normal");
  // formatDateTimeRelative("datetime-relative");
  // submitNewFeedback();
});

const triggerFlashModal = () => {
  const flashModalBodyElement = document.getElementById("flashModalBody");
  if (flashModalBodyElement.children.length > 0) {
    const modal = bootstrap.Modal.getOrCreateInstance(document.getElementById("flashModal"));
    modal.show();
  }
};

const clearFlashModalOnHidden = () => {
  const flashModalElement = document.getElementById("flashModal");
  const flashModal = bootstrap.Modal.getOrCreateInstance(flashModalElement);
  const flashModalBodyElement = document.getElementById("flashModalBody");

  flashModalElement.addEventListener('hidden.bs.modal', function handler() {
      flashModalElement.removeEventListener('hidden.bs.modal', handler);
      flashModalBodyElement.innerHTML = ``;
  });


  // if (!flashModalElement.dataset.listenerAttached) {
  //   flashModalElement.addEventListener("hidden.bs.modal", () => {
  //     const flashModalBodyElement = document.getElementById("flashModalBody");
  //     if (flashModalBodyElement.children.length > 0) {
  //       flashModalBodyElement.innerHTML = "";
  //     }
  //   });

  //   flashModalElement.dataset.listenerAttached = "true";
  // }
};

// const submitNewFeedbackOLD = () => {
//   const formElem = document.getElementById("addFeedbackModalForm");
//   if (formElem) {
//     formElem.addEventListener("submit", async function (event) {
//       event.preventDefault();
//       url = event.target.action;
//       slug = event.target.dataset.boardSlug;

//       errElem = document.getElementById("addFeedbackErrorAlert");
//       errElem.classList.add("d-none");
//       errElem.textContent = "";

//       csrfElem = document.getElementById("addFeedbackCSRF");
//       csrf = csrfElem.value.trim() || "";
//       if (csrf == "") {
//         errElem.classList.remove("d-none");
//         errElem.textContent =
//           "CSRF is invalid. Try refreshing the page before adding a feedback";
//         return;
//       }

//       titleElem = document.getElementById("addFeedbackTitle");
//       title = titleElem.value.trim() || "";
//       if (title == "") {
//         errElem.classList.remove("d-none");
//         errElem.textContent = "Title is required";
//         return;
//       }

//       descElem = document.getElementById("addFeedbackDescription");
//       desc = descElem.value.trim() || "";

//       email = "";
//       emailElem = document.getElementById("addFeedbackEmail");
//       if (emailElem) {
//         email = emailElem.value;
//       }

//       disableAddFeedbackModalElements();

//       const res = await fetch(url, {
//         method: "POST",
//         headers: {
//           "Content-Type": "application/json",
//           "X-CSRF-Token": csrf,
//         },
//         body: JSON.stringify({
//           title: title,
//           description: desc,
//           email: email,
//           boardSlug: slug,
//         }),
//       });

//       const data = await res.json();
//       enableAddFeedbackModalElements();
//       if (!res.ok) {
//         errElem = document.getElementById("addFeedbackErrorAlert");
//         errElem.classList.remove("d-none");
//         errElem.textContent = data.payload.message;
//         return;
//       }

//       clearAddFeedbackModalElements();
//       const modalEl = document.getElementById("addFeedbackModal");

//       const modal = bootstrap.Modal.getInstance(modalEl);
//       modal.hide();

//       showToastSuccess(data.payload.message);
//     });
//   }
// };

// const disableAddFeedbackModalElements = () => {
//   elementList = [
//     document.getElementById("addFeedbackTitle"),
//     document.getElementById("addFeedbackDescription"),
//     document.getElementById("addFeedbackEmail"),
//     document.getElementById("addFeedbackModalXBtn"),
//     document.getElementById("addFeedbackModalCancelBtn"),
//     document.getElementById("addFeedbackModalSubmitBtn"),
//   ];

//   elementList.forEach((element) => {
//     if (element) {
//       element.disabled = true;
//     }
//   });
// };

// const enableAddFeedbackModalElements = () => {
//   elementList = [
//     document.getElementById("addFeedbackTitle"),
//     document.getElementById("addFeedbackDescription"),
//     document.getElementById("addFeedbackEmail"),
//     document.getElementById("addFeedbackModalXBtn"),
//     document.getElementById("addFeedbackModalCancelBtn"),
//     document.getElementById("addFeedbackModalSubmitBtn"),
//   ];

//   elementList.forEach((element) => {
//     if (element) {
//       element.disabled = false;
//     }
//   });
// };

// const clearAddFeedbackModalElements = () => {
//   elementList = [
//     document.getElementById("addFeedbackTitle"),
//     document.getElementById("addFeedbackDescription"),
//     document.getElementById("addFeedbackEmail"),
//   ];

//   elementList.forEach((element) => {
//     if (element) {
//       element.value = "";
//     }
//   });
// };

// const showToastSuccess = (message) => {
//   document.getElementById("feedbackToast").textContent = message;
//   const toastEl = document.getElementById("myToast");
//   const toast = bootstrap.Toast.getOrCreateInstance(toastEl);
//   toast.show();
// };

const formatDateTimeNormal = (className) => {
  const { DateTime } = luxon;

  const elems = document.querySelectorAll("." + className);
  elems.forEach((el) => {
    const raw = el.textContent.trim();
    const dt = DateTime.fromISO(raw);
    el.textContent = dt.toFormat("dd LLL yyyy");
  });
};

// const formatDateTimeRelative = (className) => {
//   const { DateTime } = luxon;

//   const elems = document.querySelectorAll("." + className);
//   elems.forEach((el) => {
//     const raw = el.textContent.trim();
//     const dt = DateTime.fromISO(raw);
//     el.textContent = dt.toRelative();
//   });
// };

const fetchFeedbackDetailsModal = async (event, btn) => {
  const pendingModalElem = document.getElementById("pendingFeedbackViewModal");
  const pendingModal = bootstrap.Modal.getOrCreateInstance(pendingModalElem);
  pendingModal.show();
  const slug = btn.dataset.slug;

  try {
    const res = await fetch("/api/feedbacks/" + slug);
    const data = await res.json();

    if (!res.ok) {
      console.log(res.status);
      pendingModal.hide();
      return;
    }

    document.getElementById("pendingFeedbackTitlePModal").textContent =
      data.payload.title;

    if(data.payload.description.length > 0){
      document.getElementById("pendingFeedbackDescriptionPModal").textContent = data.payload.description;
    }else{
      document.getElementById("pendingFeedbackDescriptionPModal").innerHTML = `<em>No Description</em>`;
    }

    document.getElementById("modalHeaderDiv").style.display = "";
    document.getElementById("loadingDiv").style.display = "none";
    document.getElementById("modalContentDiv").style.display = "";
  } catch (err) {
    console.error(err);
    pendingModal.hide();
  }
};

const cleanPendingFeedbackViewModal = () => {
  const pendingModalElem = document.getElementById("pendingFeedbackViewModal");

  if (pendingModalElem) {
    pendingModalElem.addEventListener("hidden.bs.modal", () => {
      document.getElementById("pendingFeedbackTitlePModal").textContent = "";
      document.getElementById("pendingFeedbackDescriptionPModal").textContent =
        "";
      document.getElementById("pendingFeedbackDescriptionPModal").innerHTML =
      "";

      document.getElementById("modalHeaderDiv").style.display = "none";
      document.getElementById("loadingDiv").style.display = "";
      document.getElementById("modalContentDiv").style.display = "none";
    });
  }
};


/**
 * Add feedback submission using the modal
 * @param {*} event 
 * @param {*} form 
 * @returns 
 */
const submitNewFeedback = async (event, form) => {
  
  // ---- prevent the default submission flow ---- //
  event.preventDefault();

  // ---- first remove any error messages --- //
  // from general
  genErrorElem = document.getElementById("addFeedbackErrorAlert");
  if (genErrorElem){
    genErrorElem.textContent = "";
    if (!genErrorElem.classList.contains('d-none')) {
      genErrorElem.classList.add("d-none")
    }
    
  }

  // from title
  titleErrorElem = document.getElementById("addFeedbackTitleError");
  if (titleErrorElem){
    titleErrorElem.textContent = ""
    if (!titleErrorElem.classList.contains('d-none')) {
      titleErrorElem.classList.add("d-none")
    }
  }


  // ----- now get all the relevant data ----- //
  // get csrf
  csrf = ""
  csrfElem = document.getElementById("addFeedbackCSRF");
  if (csrfElem){
    csrf = csrfElem.value;
  }

  // get board slug from the action data attribute
  slug = form.dataset.boardSlug;


  // get title
  title = ""
  titleElem = document.getElementById("addFeedbackTitle");
  if (titleElem){
    title = titleElem.value;
  }


  // get description
  description = ""
  descElem = document.getElementById("addFeedbackDescription");
  if (descElem){
    description = descElem.value;
  }


  // check title is within range

  if(title.length < 3 || title.length > 60){
    titleErrorElem.textContent = "Must be between 3 and 60 characters long";

    if (titleErrorElem.classList.contains('d-none')) {
      titleErrorElem.classList.remove("d-none")
    }
    return;
  }

  // -------------------- Disable Elements For Submission ------------ //
  const submitBtn = document.getElementById("addFeedbackModalSubmitBtn");
  const cancelBtn = document.getElementById("addFeedbackModalCancelBtn");
  const xBtn = document.getElementById("addFeedbackModalXBtn");
  titleElem.disabled = true
  csrfElem.disabled = true
  descElem.disabled = true
  submitBtn.disabled = true
  cancelBtn.disabled = true
  xBtn.disabled = true

  submitBtn.textContent = "Submitting..."


  try{
    const resp = await fetch(form.action, {
      method: "POST",
      headers: {
        "Content-Type":"application/json",
        "X-CSRF-Token": csrf,
      },
      body: JSON.stringify({
        title: title,
        description: description,
        boardSlug: slug
      })
    });

    if (!resp.ok){
      throw new Error(`HTTP error: ${resp.status}`);
    }

    const result = await resp.json();

    // ------------ Close Add Feedback Modal and Show Flash Modal with Message ------------ //

    // get the add feedback modal
    const addFeedbackModalElem = document.getElementById("addFeedbackModal");
    const addFeedbackModal = bootstrap.Modal.getOrCreateInstance(addFeedbackModalElem);

    // get flash modal
    const flashModalElem = document.getElementById("flashModal");
    const flashModal = bootstrap.Modal.getOrCreateInstance(flashModalElem);
    const flashModalBodyElem = document.getElementById("flashModalBody")
    flashModalBodyElem.innerHTML = `<div class="alert alert-primary" role="alert">${result.payload.message}</div>`


    addFeedbackModalElem.addEventListener('hidden.bs.modal', function handler() {
        addFeedbackModalElem.removeEventListener('hidden.bs.modal', handler);
        flashModal.show();
    });

    if (addFeedbackModal && addFeedbackModalElem.classList.contains('show')) {
      addFeedbackModal.hide();
    }

    // ------ Submitted successfully, clear content ------- //
    titleElem.value = ""
    descElem.value = ""


  }catch(error){
    if (genErrorElem){
      genErrorElem.textContent = error;
      if (genErrorElem.classList.contains('d-none')) {
        genErrorElem.classList.remove("d-none")
      }
      
    }
  }finally{
    // Re-enable elements
    titleElem.disabled = false
    csrfElem.disabled = false
    descElem.disabled = false
    submitBtn.disabled = false
    cancelBtn.disabled = false
    xBtn.disabled = false

    submitBtn.textContent = "Submit"
  }

}



/**
 * Get feedback using api
 * @param {*} event 
 * @param {*} elem 
 */
const getFeedback = async (event, elem) => {
  const slug = elem.dataset.slug;

  const feedbackModalElem = document.getElementById("feedbackModal");
  const feedbackModal = bootstrap.Modal.getOrCreateInstance(feedbackModalElem);

  feedbackModal.show();

  const errorDivElem = document.getElementById("feedbackModalErrorDiv");
  const errorDivTextElem = document.getElementById("feedbackModalErrorTextDiv");
  const loadingDivElem = document.getElementById("feedbackModalLoadingDiv");
  const dataDivElem = document.getElementById("feedbackModalDataDiv");

  const titleDivElem = document.getElementById("feedbackModalDataTitleDiv");
  const descDivElem = document.getElementById("feedbackModalDataDescDiv");
  const createdAtDivElem = document.getElementById("feedbackModalDataCreatedAtDiv");
  const reviewDivElem = document.getElementById("feedbackModalDataReviewDiv");
  const reviewAnchorElem = document.getElementById("feedbackModalDataReviewAnchor");
  const scopeSpanElem = document.getElementById("feedbackModalScope");
  const resolutionSpanElem = document.getElementById("feedbackModalResolution");
  const reportAnchorElem = document.getElementById("feedbackModalReport");

  // remove any existing data
  // error
  errorDivTextElem.textContent = ""

  // title
  titleDivElem.textContent = ""
  // description
  descDivElem.textContent = ""
  descDivElem.innerHTML = ""

  // anchor
  reviewAnchorElem.href = ""
  reportAnchorElem.href = ""

  // created
  createdAtDivElem.textContent = ""

  // scope
  scopeSpanElem.textContent = ""
  // resolution
  resolutionSpanElem.textContent = ""



  // hide the error div element
  if (!errorDivElem.classList.contains('d-none')) {
    errorDivElem.classList.add("d-none")
  }

  // hide the data div element
  if (!dataDivElem.classList.contains('d-none')) {
    dataDivElem.classList.add("d-none")
  }

  if (!reviewDivElem.classList.contains('d-none')) {
    reviewDivElem.classList.add("d-none")
  }

  // show the loading div element
  if (loadingDivElem.classList.contains('d-none')) {
    loadingDivElem.classList.remove("d-none")
  }

  

  try {
    const resp = await fetch("/api/feedbacks/" + slug, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    });

    const result = await resp.json();

    if (!resp.ok) {
      throw new Error(`HTTP error: ${resp.status}`);
    }

    const payload = result.payload;


    dataDivElem.classList.remove("d-none")
    loadingDivElem.classList.add("d-none")
    if(payload.is_board_owner){
      reviewDivElem.classList.remove("d-none")
      reviewAnchorElem.href  = "/feedbacks/review?f="+payload.slug
    }
    reportAnchorElem.href = "/report/feedbacks?f="+payload.slug;


    titleDivElem.textContent = payload.title
    if(payload.description.length > 0){
      descDivElem.textContent = payload.description;
    }else{
      descDivElem.innerHTML = `<span class="text-info"><em>No Description</em></span>`
    }
    createdAtDivElem.textContent = formatDefault(payload.created_at);
    scopeSpanElem.textContent = payload.scope;
    resolutionSpanElem.textContent = payload.resolution;


  } catch (error) {
    errorDivTextElem.textContent = error
    // show the error div element
    if (errorDivElem.classList.contains("d-none")) {
      errorDivElem.classList.remove("d-none");
    }

    // hide the data div element
    if (!dataDivElem.classList.contains("d-none")) {
      dataDivElem.classList.add("d-none");
    }

    // hide the loading div element
    if (!loadingDivElem.classList.contains("d-none")) {
      loadingDivElem.classList.add("d-none");
    }
  }
}




// ======== Date Utilities ============ //
const formatDateTimeRelative = () => {
  const { DateTime } = luxon;

  const elems = document.querySelectorAll(".dt-relative");

  elems.forEach((el) => {
    const raw = el.textContent.trim();
    const dt = DateTime.fromISO(raw, { zone: "utc" });

    if (!dt.isValid) {
      console.warn("Invalid date:", raw);
      return;
    }

    el.textContent = dt.toRelative();
  });
}

const formatDateTimeDefault = () => {
  const elems = document.querySelectorAll(".dt-default");

  elems.forEach((el) => {
    const raw = el.textContent.trim();
    el.textContent = formatDefault(raw)
  });
}



const formatDefault = s => {
  const { DateTime } = luxon;

  const dt = DateTime.fromISO(s, { zone: "utc" });
  if (!dt.isValid) {
    console.warn("Invalid date:", s);
    return;
  }

  return dt.toFormat("dd LLL yyyy HH:MM");
}

const formatDateDefault = () => {
  const { DateTime } = luxon;

  const elems = document.querySelectorAll(".d-default");

  elems.forEach((el) => {
    const raw = el.textContent.trim();
    const dt = DateTime.fromISO(raw, { zone: "utc" });

    if (!dt.isValid) {
      console.warn("Invalid date:", raw);
      return;
    }

    el.textContent = dt.toFormat("dd LLL yyyy");
  });
}


// ======== String Utilities ============ //
function capitalizeWords(str) {
  return str
    .split(" ")
    .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(" ");
}



// =============== Filter In Board Feature ============ //

const applyFilter = () => {
  const url = new URL(window.location.href);
  pageParam = url.searchParams.get("page")
  if(pageParam == null){
    pageParam = 1;
  }

  const pathName = window.location.pathname;

  scope = "all"
  const scopeRadio = document.querySelector('input[name="filterScopeRadio"]:checked');
  if (scopeRadio) {
    scope = scopeRadio.value
  }

  resolution = "all"
  const resolutionRadio = document.querySelector('input[name="filterResolutionRadio"]:checked');
  if (resolutionRadio) {
    resolution = resolutionRadio.value
    
  }


  const params = new URLSearchParams();
  params.set("scope", scope)
  params.set("resolution", resolution)
  params.set("page", pageParam)

  const fullPathWithQuery = pathName+"?"+params.toString();
  window.location.replace(fullPathWithQuery)

  
}