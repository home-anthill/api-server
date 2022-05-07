export function getHeaders() {
  let token = localStorage.getItem('token');
  return {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer ' + token
  };
}
