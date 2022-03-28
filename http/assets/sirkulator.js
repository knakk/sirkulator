'use strict';

function ready(fn) {
    if (document.readyState !== "loading"){
      fn();
    } else {
      document.addEventListener("DOMContentLoaded", fn);
    }
}

function initSearchSelect() {
    function removeSearchSelected(event) {
        const li = event.target.closest("li");
        const value = li.children[1].value;
        const label = li.children[0].innerText;

        // Add back as an option to datalist before removing
        // TODO sort options list
        const option = new Option(label, value);
        event.target.closest("div.field").querySelector("datalist").appendChild(option);

        // Enable input if disabled
        const el = event.target.closest("div.field").querySelector("input.single-value")
        if (el) {
            el.disabled = false;
        }

        li.remove();
    };

    function searchSelectVal(target, value, label, idx) {
        const li = document.createElement("li")
        li.innerHTML = `
            <div class="selected-term">${label}</div><input type="hidden" name="${target.id}" value="${value}"><button type="button" class="unselect-term"><span>✕</span></button>`;
        li.addEventListener("click", (event) => removeSearchSelected(event));
        target.parentElement.querySelector("ul").appendChild(li);
        target.list.options[idx].remove();
        target.value="";

        // Disabled input if single-value
        const el = target.closest("div.field").querySelector("input.single-value")
        if (el) {
            el.disabled = true;
        }
    }

    function searchSelectVerify(event) {
        let value = event.target.value;
        let label;
        let idx;
        for (let i=0; i < event.target.list.options.length; i++) {
            if (event.target.list.options[i].value.toLowerCase() === value.toLowerCase() ||
                event.target.list.options[i].text.toLowerCase() === value) {
                label = event.target.list.options[i].text;
                value = event.target.list.options[i].value; // in-case we have different lowercase/uppercase val
                idx = i;
                break;
            }
        }
        // TODO maybe also select if only one element in list visible?
        if (label) {
            searchSelectVal(event.target, value, label, idx);
        }
    }

    const els = document.getElementsByClassName("search-select");

    for (let i=0; i < els.length; i++) {
        els[i].addEventListener("keyup", (event) => {
            if (event.which === 13) {
                searchSelectVerify(event);
            }
        });
        els[i].addEventListener("change", (event) => {
            searchSelectVerify(event);
        });
    }

    document.querySelectorAll(".unselect-term").forEach(
        el => el.addEventListener("click", (event) => removeSearchSelected(event)));

}

ready(function() {
    initSearchSelect();

    document.body.addEventListener('htmx:afterSwap', function() {
        initSearchSelect();
    });
});
