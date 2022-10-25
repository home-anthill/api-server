from setuptools import setup

setup(
    name='esp32-configurator',
    version='1.0.0',
    packages=['configurator'],
    entry_points={
        'console_scripts': [
            'esp32-configurator = configurator.__main__:main'
        ]
    })