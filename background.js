/*
 * messages sent from the external server will have the `action`
 * property (a string) and a `content` object to be unpacked in
 * in the BrowserTabs method specified by `action`
 */

function handleMsg(msg) {
  console.log("Received:", msg)
  if (!msg.action) {
    port.postMessage({
      action: "error",
      clientId: msg.clientId,
      content: "Received msg with no action"
    })
    return
  }

  switch (msg.action) {

    case "query":
      browser.tabs.query(msg.content).then(
        (result) => {
          for (const t of result) {
            response = {
              action: "created",
              clientId: msg.clientId,
              content: t,
            }
            console.log("SENDING: ", response)
            port.postMessage(response)
          }
        },
        (error) => {}
      )

    default:
      port.postMessage({
        action: "error",
        clientId: msg.clientId,
        content: `Action ${msg.Action} is unknown`
      })
  }
// captureTab
// captureVisibleTab
// connect
// detectLanguage
// executeScript
// get
// getAllInWindow
// getCurrent
// getSelected
// getZoom
// getZoomSettings
}

var port

function reconnect() {
  if (port) {
    if (browser.runtime.lastError) {
      console.log("DISCONNECT", browser.runtime.lastError)
    } else {
      console.log("DISCONNECT: unknown reason")
    }
    console.log("Attempting restart")
  } else {
    console.log("Starting gateway")
  }
  port = browser.runtime.connectNative("tabs_server")
  port.onDisconnect.addListener(
    (p) => {reconnect()}
  )
  port.onMessage.addListener((msg) => { handleMsg(msg) })
  return port
}

port = reconnect()


/* activeInfo {tabId, previousTabId, windowId} */
browser.tabs.onActivated.addListener(
  (activeInfo) => {
    console.log("Tab activated:", activeInfo)
    port.postMessage({
      action: "activated",
      content: {
        tabId: activeInfo.tabId,
        previous: activeInfo.previousTabId,
        windowId: activeInfo.windowId
      }
    })
  }
)

/* changeInfo contains the tab properties that changed */
browser.tabs.onUpdated.addListener(
  (tabId, changeInfo, tab) => {
    console.log("Tab updated:", tabId, changeInfo)
    port.postMessage({
      action: "updated",
      content: {
        tabId: tabId,
        delta: changeInfo,
      }
    })
  }
)

browser.tabs.onCreated.addListener(
  (tab) => {
    console.log("Tab created:", tab)
    port.postMessage({
      action: "created",
      content: tab,
    })
  }
)

/* moveInfo {windowId, fromIndex, toIndex} */
browser.tabs.onMoved.addListener(
  (tabId, moveInfo) => {
    console.log("Tab moved:", tabId, moveInfo)
    port.postMessage({
      action: "moved",
      content: {
        tabId: tabId,
        windowId: moveInfo.windowId,
        fromIndex: moveInfo.fromIndex,
        toIndex: moveInfo.toIndex,
      }
    })
  }
)

/* removeInfo {windowId, isWindowClosing} */
browser.tabs.onRemoved.addListener(
  (tabId, removeInfo) => {
    console.log("Tab removed:", tabId, removeInfo)
    port.postMessage({
      action: "removed",
      content: {
        tabId: tabId,
        windowId: removeInfo.windowId,
        isWindowClosing: removeInfo.isWindowClosing,
      }
    })
  }
)

/* attachInfo {newWindowId, newPosition} */
browser.tabs.onAttached.addListener(
  (tabId, info) => {
    console.log("Tab attached:", tabId, info)
    port.postMessage({
      action: "attached",
      content: {
        tabId: tabId,
        windowId: info.oldWindowId,
        position: info.oldPosition,
      }
    })
  }
)

/* detachInfo {oldWindowId, oldPosition} */
browser.tabs.onDetached.addListener(
  (tabId, info) => {
    console.log("Tab detached:", tabId, info)
    port.postMessage({
      action: "detached",
      content: {
        tabId: tabId,
        windowId: info.oldWindowId,
        position: info.oldPosition,
      }
    })
  }
)


/*

// highlightInfo {windowId, tabIds}
browser.tabs.onHighlighted.addListener(
  (highlightInfo) => {

  }
)

// zoomChangeInfo {tabId, oldZoomFactor, newZoomFactor, zoomSettings}
browser.tabs.onZoomChange.addListener(
  (zoomChangeInfo) => {

  }
)

*/
