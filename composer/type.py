import dataclasses
from typing import List

@dataclasses.dataclass
class ScriptItem:
    cn: str
    en: str
    image_prompt: str
    voice_path: str
    image_path: str = ''

@dataclasses.dataclass
class MovieMeta:
    title: str
    script_items: List[ScriptItem]
    workdir: str = ''

