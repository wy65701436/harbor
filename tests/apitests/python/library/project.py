# -*- coding: utf-8 -*-

import sys
import base
import swagger_client

class Project(base.Base):
    def create_project(self, name=base._random_name("project"), 
        metadata = {}, **kwargs):
        print "***********"
        print name
        client = self._get_client(**kwargs)
        _, _, header = client.projects_post_with_http_info(
            swagger_client.ProjectReq(name, metadata))
        return base._get_id_from_header(header), name