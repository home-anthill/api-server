from pydantic import BaseModel


class Secrets(BaseModel):
    wifi_ssid: str
    wifi_password: str

    manufacturer: str = 'ks89'
    api_token: str

    ssl: bool = True

    server_domain: str
    server_port: str = '443'
    server_path: str = '/api/register'

    mqtt_domain: str
    mqtt_port: str = 8883
