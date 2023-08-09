export const backendUrl = "/api"; // for production
//export const backendUrl = "http://localhost:3000"; // for development
export function timeAgo(d) {
    let t = (Date.now() - new Date(d)) / 1000, r = " ago";
    return t < 60 ? `${t | 0}s`+r : t < 3600 ? `${(t / 60) | 0}m`+r : t < 86400 ? `${(t / 3600) | 0}h`+r : `${(t / 86400) | 0}d`+r;
}