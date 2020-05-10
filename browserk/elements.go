package browserk

type HTMLElementType int16

// revive:disable:var-naming
const (
	// Main Root
	HTML HTMLElementType = iota

	// METADATA
	BASE
	HEAD
	LINK
	META
	STYLE
	TITLE

	// Sectioning Root
	BODY

	//Content
	ADDRESS
	ARTICLE
	ASIDE
	FOOTER
	HEADER
	H1
	H2
	H3
	H4
	H5
	H6
	HGROUP
	MAIN
	NAV
	SECTION

	// Text
	BLOCKQUOTE
	DD
	DIV
	DL
	DT
	FIGCAPTION
	FIGURE
	HR
	LI
	// MAIN
	OL
	P
	PRE
	UL

	// Inline Text
	A
	ABBR
	B
	BDI
	BDO
	BR
	CITE
	CODE
	DATA
	DFN
	EM
	I
	KBD
	MARK
	Q
	RB
	RP
	RT
	RTC
	RUBY
	S
	SAMP
	SMALL
	SPAN
	STRONG
	SUB
	SUP
	TIME
	U
	VAR
	WBR

	// Image and Multimedia
	AREA
	AUDIO
	IMG
	MAP
	TRACK
	VIDEO

	// Embedded
	EMBED
	IFRAME
	OBJECT
	PARAM
	PICTURE
	SOURCE

	// Scripting
	CANVAS
	NOSCRIPT
	SCRIPT

	// Demarcating
	DEL
	INS

	// Table Content
	CAPTION
	COL
	COLGROUP
	TABLE
	TBODY
	TD
	TFOOT
	TH
	THEAD
	TR

	// Forms
	BUTTON
	DATALIST
	FIELDSET
	FORM
	INPUT
	LABEL
	LEGEND
	METER
	OPTGROUP
	OPTION
	OUTPUT
	PROGRESS
	SELECT
	TEXTAREA

	// Interactive
	DETAILS
	DIALOG
	MENU
	SUMMARY

	// Web Components
	SLOT
	TEMPLATE

	// Deprecated
	ACRONYM
	APPLET
	BASEFONT
	BGSOUND
	BIG
	BLINK
	CENTER
	COMMAND
	CONTENT
	DIR
	ELEMENT
	FONT
	FRAME
	FRAMESET
	IMAGE
	ISINDEX
	KEYGEN
	LISTING
	MARQUEE
	MENUITEM
	MULTICOL
	NEXTID
	NOBR
	NOEMBED
	NOFRAMES
	PLAINTEXT
	SHADOW
	SPACER
	STRIKE
	TT
	XMP
)

type HTMLEventType int8

// revive:exported
const (
	// Resource events
	HtmlEvterror        HTMLEventType = iota //  A resource failed to load.
	HtmlEvtabort                             //  The loading of a resource has been aborted.
	HtmlEvtload                              //  A resource and its dependent resources have finished loading.
	HtmlEvtbeforeunload                      //  The window, the document and its resources are about to be unloaded.
	HtmlEvtunload                            //  The document or a dependent resource is being unloaded.

	// focus events
	HtmlEvtfocus    // An element has received focus (does not bubble).
	HtmlEvtblur     // An element has lost focus (does not bubble).
	HtmlEvtfocusin  // An element is about to receive focus (does bubble).
	HtmlEvtfocusout // An element is about to lose focus (does bubble).

	// css events
	HtmlEvtanimationstart     // A CSS animation has started.
	HtmlEvtanimationcancel    // A CSS animation has aborted.
	HtmlEvtanimationend       // A CSS animation has completed.
	HtmlEvtanimationiteration // A CSS animation is repeated.
	HtmlEvttransitionstart    // A CSS transition has actually started (fired after any delay).
	HtmlEvttransitioncancel   // A CSS transition has been cancelled.
	HtmlEvttransitionend      // A CSS transition has completed.
	HtmlEvttransitionrun      // A CSS transition has begun running (fired before any delay starts).

	//  form events
	HtmlEvtreset  // The reset button is pressed
	HtmlEvtsubmit // The submit button is pressed

	//  print events
	HtmlEvtbeforeprint // The print dialog is opened
	HtmlEvtafterprint  // The print dialog is closed

	//  text composition events
	HtmlEvtcompositionstart  // The composition of a passage of text is prepared (similar to keydown for a keyboard input, but works with other inputs such as speech recognition).
	HtmlEvtcompositionupdate // A character is added to a passage of text being composed.
	HtmlEvtcompositionend    // The composition of a passage of text has been completed or canceled.

	// view events
	HtmlEvtfullscreenchange // An element was toggled to or from fullscreen mode.
	HtmlEvtfullscreenerror  // It was impossible to switch to fullscreen mode for technical reasons or because the permission was denied.
	HtmlEvtresize           // The document view has been resized.
	HtmlEvtscroll           // The document view or an element has been scrolled.

	//  clipboard events
	HtmlEvtcut   // The selection has been cut and copied to the clipboard
	HtmlEvtcopy  // The selection has been copied to the clipboard
	HtmlEvtpaste // The item from the clipboard has been pasted

	// keyboard events
	HtmlEvtkeydown  // ANY key is pressed
	HtmlEvtkeypress // ANY key (except Shift, Fn, or CapsLock) is in pressed position. (Fired continously.)
	HtmlEvtkeyup    // ANY key is released

	//  mouse events
	HtmlEvtauxclick          // A pointing device button (ANY non-primary button) has been pressed and released on an element.
	HtmlEvtclick             // A pointing device button (ANY button; soon to be primary button only) has been pressed and released on an element.
	HtmlEvtcontextmenu       // The right button of the mouse is clicked (before the context menu is displayed).
	HtmlEvtdblclick          // A pointing device button is clicked twice on an element.
	HtmlEvtmousedown         // A pointing device button is pressed on an element.
	HtmlEvtmouseenter        // A pointing device is moved onto the element that has the listener attached.
	HtmlEvtmouseleave        // A pointing device is moved off the element that has the listener attached.
	HtmlEvtmousemove         // A pointing device is moved over an element. (Fired continously as the mouse moves.)
	HtmlEvtmouseover         // A pointing device is moved onto the element that has the listener attached or onto one of its children.
	HtmlEvtmouseout          // A pointing device is moved off the element that has the listener attached or off one of its children.
	HtmlEvtmouseup           // A pointing device button is released over an element.
	HtmlEvtpointerlockchange // The pointer was locked or released.
	HtmlEvtpointerlockerror  // It was impossible to lock the pointer for technical reasons or because the permission was denied.
	HtmlEvtselect            // Some text is being selected.
	HtmlEvtwheel             // A wheel button of a pointing device is rotated in any direction.

	//  drag & drop events
	HtmlEvtdrag      // An element or text selection is being dragged. (Fired continuously every 350ms)
	HtmlEvtdragend   // A drag operation is being ended (by releasing a mouse button or hitting the escape key).
	HtmlEvtdragenter // A dragged element or text selection enters a valid drop target.
	HtmlEvtdragstart // The user starts dragging an element or text selection.
	HtmlEvtdragleave // A dragged element or text selection leaves a valid drop target.
	HtmlEvtdragover  // An element or text selection is being dragged over a valid drop target. (Fired continuously every 350ms)
	HtmlEvtdrop      // An element is dropped on a valid drop target.

	// media events
	HtmlEvtaudioprocess   // The input buffer of a ScriptProcessorNode is ready to be processed.
	HtmlEvtcanplay        // The browser can play the media, but estimates that not enough data has been loaded to play the media up to its end without having to stop for further buffering of content.
	HtmlEvtcanplaythrough // The browser estimates it can play the media up to its end without stopping for content buffering.
	HtmlEvtcomplete       // The rendering of an OfflineAudioContext is terminated.
	HtmlEvtdurationchange // The duration attribute has been updated.
	HtmlEvtemptied        // The media has become empty; for example, this event is sent if the media has already been loaded (or partially loaded), and the load() method is called to reload it.
	HtmlEvtended          // Playback has stopped because the end of the media was reached.
	HtmlEvtloadeddata     // The first frame of the media has finished loading.
	HtmlEvtloadedmetadata // The metadata has been loaded.
	HtmlEvtpause          // Playback has been paused.
	HtmlEvtplay           // Playback has begun.
	HtmlEvtplaying        // Playback is ready to start after having been paused or delayed due to lack of data.
	HtmlEvtratechange     // The playback rate has changed.
	HtmlEvtseeked         // A seek operation completed.
	HtmlEvtseeking        // A seek operation began.
	HtmlEvtstalled        // The user agent is trying to fetch media data, but data is unexpectedly not forthcoming.
	HtmlEvtsuspend        // Media data loading has been suspended.
	HtmlEvttimeupdate     // The time indicated by the currentTime attribute has been updated.
	HtmlEvtvolumechange   // The volume has changed.
	HtmlEvtwaiting        // Playback has stopped because of a temporary lack of data.

	// progress events
	HtmlEvtloadend   // Progress has stopped (after "error", "abort", or "load" have been dispatched).
	HtmlEvtloadstart // Progress has begun.
	HtmlEvtprogress  // In progress.
	HtmlEvttimeout   // Progression is terminated due to preset time expiring.

	// touch events
	HtmlEvttouchcancel
	HtmlEvttouchend
	HtmlEvttouchmove
	HtmlEvttouchstart

	// pointer events
	HtmlEvtpointerover
	HtmlEvtpointerenter
	HtmlEvtpointerdown
	HtmlEvtpointermove
	HtmlEvtpointerup
	HtmlEvtpointercancel
	HtmlEvtpointerout
	HtmlEvtpointerleave
	HtmlEvtgotpointercapture
	HtmlEvtlostpointercapture
)

// HTMLElement type and importance
type HTMLElement struct {
	Type       HTMLElementType
	Weight     float32
	Location   string
	Attributes map[string]string
}
