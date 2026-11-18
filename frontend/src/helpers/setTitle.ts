export function setTitle(title : undefined | string) {
	document.title = (typeof title === 'undefined' || title === '')
		? 'Central de Suporte'
		: `${title} | Central de Suporte`
}
